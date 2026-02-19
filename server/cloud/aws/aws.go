package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	schedulertypes "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/daylamtayari/cierge/reservation"
	"github.com/daylamtayari/cierge/server/cloud"
	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
)

var (
	ErrDecodeConfig           = errors.New("failed to decode config")
	ErrMissingRegion          = errors.New("region is required")
	ErrMissingKMSKeyID        = errors.New("KMS key ID is required")
	ErrMissingLambdaARN       = errors.New("lambda ARN is required")
	ErrMissingSchedulerRole   = errors.New("scheduler role ARN is required")
	ErrMissingColdStartBuffer = errors.New("cold start buffer is required")
	ErrInvalidColdStartBuffer = errors.New("cold start buffer is not a valid duration")
	ErrInvalidKMSKeyID        = errors.New("KMS key ID value is invalid")
	ErrInvalidLambdaARN       = errors.New("lambda ARN value is invalid")
	ErrInvalidSchedulerRole   = errors.New("schedule role ARN is invalid")
	ErrMissingCredentials     = errors.New("no credentials found in environment or config")
	ErrCredentialValidation   = errors.New("invalid credentials provided")
)

const scheduleNamePrefix = "cierge-job-"

type Provider struct {
	scheduler       *scheduler.Client
	kms             *kms.Client
	lambdaARN       string
	roleARN         string
	kmsKeyID        string
	coldStartBuffer time.Duration
}

// AWS provider configuration
type providerConfig struct {
	Region           string `json:"region"`
	KMSKeyID         string `json:"kms_key_id"`
	LambdaARN        string `json:"lambda_arn"`
	SchedulerRoleARN string `json:"scheduler_role_arn"`
	ColdStartBuffer  string `json:"cold_start_buffer"`
	AccessKeyID      string `json:"access_key_id"`
	SecretAccessKey  string `json:"secret_access_key"`
	SessionToken     string `json:"session_token"`
}

// Creates a new AWS provider
func NewProvider(cfg map[string]any) (cloud.Provider, error) {
	pCfg, err := decodeConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Value has already been validated
	coldStartBuffer, _ := time.ParseDuration(pCfg.ColdStartBuffer)

	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(pCfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(pCfg.AccessKeyID, pCfg.SecretAccessKey, pCfg.SessionToken)),
	)
	if err != nil {
		return nil, err
	}

	return &Provider{
		scheduler:       scheduler.NewFromConfig(awsCfg),
		kms:             kms.NewFromConfig(awsCfg),
		lambdaARN:       pCfg.LambdaARN,
		roleARN:         pCfg.SchedulerRoleARN,
		kmsKeyID:        pCfg.KMSKeyID,
		coldStartBuffer: coldStartBuffer,
	}, nil
}

// Use EventBridge to schedule a one-time execution of the reservation lambda
// with the reservation event as the payload
func (p *Provider) ScheduleJob(ctx context.Context, event reservation.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	name := scheduleNamePrefix + event.JobID.String()
	scheduledAt := event.DropTime.Add(-p.coldStartBuffer)
	expression := "at(" + scheduledAt.UTC().Format("2006-01-02T15:04:05") + ")"
	payloadStr := string(payload)

	_, err = p.scheduler.CreateSchedule(ctx, &scheduler.CreateScheduleInput{
		Name:               &name,
		ScheduleExpression: &expression,
		FlexibleTimeWindow: &schedulertypes.FlexibleTimeWindow{
			Mode: schedulertypes.FlexibleTimeWindowModeOff,
		},
		Target: &schedulertypes.Target{
			Arn:     &p.lambdaARN,
			RoleArn: &p.roleARN,
			Input:   &payloadStr,
		},
		ActionAfterCompletion: schedulertypes.ActionAfterCompletionDelete,
		KmsKeyArn:             &p.kmsKeyID,
	})
	if err != nil {
		return err
	}

	return nil
}

// Cancels an EventBridge scheduled event
func (p *Provider) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	name := scheduleNamePrefix + jobID.String()

	_, err := p.scheduler.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
		Name: &name,
	})
	if err != nil {
		var notFound *schedulertypes.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}

	return nil
}

// Updates the platform credentials for a scheduled event
func (p *Provider) UpdateJobCredentials(ctx context.Context, jobID uuid.UUID, encryptedToken string) error {
	name := scheduleNamePrefix + jobID.String()

	getOutput, err := p.scheduler.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: &name,
	})
	if err != nil {
		var notFound *schedulertypes.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return fmt.Errorf("%w: %s", cloud.ErrJobNotFound, name)
		}
		return err
	}

	var event reservation.Event
	if err := json.Unmarshal([]byte(*getOutput.Target.Input), &event); err != nil {
		return err
	}

	event.EncryptedToken = encryptedToken

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	scheduledAt := event.DropTime.Add(-p.coldStartBuffer)
	expression := "at(" + scheduledAt.UTC().Format("2006-01-02T15:04:05") + ")"
	payloadStr := string(payload)

	_, err = p.scheduler.UpdateSchedule(ctx, &scheduler.UpdateScheduleInput{
		Name:               &name,
		ScheduleExpression: &expression,
		FlexibleTimeWindow: &schedulertypes.FlexibleTimeWindow{
			Mode: schedulertypes.FlexibleTimeWindowModeOff,
		},
		Target: &schedulertypes.Target{
			Arn:     &p.lambdaARN,
			RoleArn: &p.roleARN,
			Input:   &payloadStr,
		},
		KmsKeyArn: &p.kmsKeyID,
	})
	if err != nil {
		return err
	}

	return nil
}

// Encrypts a provided string using KMS and returns the base64 encoded ciphertext
func (p *Provider) EncryptData(ctx context.Context, plaintext string) (string, error) {
	output, err := p.kms.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     &p.kmsKeyID,
		Plaintext: []byte(plaintext),
	})
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(output.CiphertextBlob), nil
}

// Decrypts a base64-encoded ciphertext using KMS and returns the plaintext
func (p *Provider) DecryptData(ctx context.Context, ciphertext string) (string, error) {
	ciphertextBlob, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	output, err := p.kms.Decrypt(ctx, &kms.DecryptInput{
		KeyId:          &p.kmsKeyID,
		CiphertextBlob: ciphertextBlob,
	})
	if err != nil {
		return "", err
	}

	return string(output.Plaintext), nil
}

// Validates an AWS config
func ValidateConfig(cfg map[string]any, isProduction bool) error {
	pCfg, err := decodeConfig(cfg)
	if err != nil {
		return err
	}

	if pCfg.Region == "" {
		return ErrMissingRegion
	}

	if pCfg.ColdStartBuffer == "" {
		return ErrMissingColdStartBuffer
	}
	if _, err := time.ParseDuration(pCfg.ColdStartBuffer); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidColdStartBuffer, pCfg.ColdStartBuffer)
	}

	if pCfg.KMSKeyID == "" {
		return ErrMissingKMSKeyID
	} else if !arn.IsARN(pCfg.KMSKeyID) {
		return ErrInvalidKMSKeyID
	}

	if pCfg.LambdaARN == "" {
		return ErrMissingLambdaARN
	} else if !arn.IsARN(pCfg.LambdaARN) {
		return ErrInvalidLambdaARN
	}

	if pCfg.SchedulerRoleARN == "" {
		return ErrMissingSchedulerRole
	} else if !arn.IsARN(pCfg.SchedulerRoleARN) {
		return ErrInvalidSchedulerRole
	}

	return validateCredentials(cfg, &pCfg)
}

// Retrieve, validate, and test credentials
// Stores the validated credentials in the cfg map
// Credential resolution: env vars -> config -> error
func validateCredentials(cfg map[string]any, pCfg *providerConfig) error {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")

	if accessKeyID == "" || secretAccessKey == "" {
		if pCfg.AccessKeyID != "" && pCfg.SecretAccessKey != "" {
			accessKeyID = pCfg.AccessKeyID
			secretAccessKey = pCfg.SecretAccessKey
			sessionToken = ""
		} else {
			return ErrMissingCredentials
		}
	}

	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(pCfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken)),
	)
	if err != nil {
		return err
	}

	// Validate credentials via STS GetCallerIdentity
	stsClient := sts.NewFromConfig(awsCfg)
	if _, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{}); err != nil {
		return fmt.Errorf("%w: %w", ErrCredentialValidation, err)
	}

	cfg["access_key_id"] = accessKeyID
	cfg["secret_access_key"] = secretAccessKey
	cfg["session_token"] = sessionToken

	return nil
}

// Decodes the config map into a struct
func decodeConfig(cfg map[string]any) (providerConfig, error) {
	var pCfg providerConfig
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &pCfg,
		TagName: "json",
	})
	if err != nil {
		return pCfg, fmt.Errorf("%w: %w", ErrDecodeConfig, err)
	}
	if err := decoder.Decode(cfg); err != nil {
		return pCfg, fmt.Errorf("%w: %w", ErrDecodeConfig, err)
	}
	return pCfg, nil
}
