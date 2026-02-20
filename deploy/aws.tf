terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.4"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "Cierge"
      ManagedBy   = "Terraform"
      Environment = var.environment
    }
  }
}

# Variables

variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev or prod)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "prod"], var.environment)
    error_message = "Environment must be dev or prod."
  }
}

variable "kms_deletion_window" {
  description = "KMS key deletion window in days"
  type        = number
  default     = 30

  validation {
    condition     = var.kms_deletion_window >= 7 && var.kms_deletion_window <= 30
    error_message = "KMS deletion window must be between 7 and 30 days."
  }
}

variable "lambda_timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 600

  validation {
    condition     = var.lambda_timeout >= 60 && var.lambda_timeout <= 900
    error_message = "Lambda timeout must be between 60 and 900 seconds."
  }
}

variable "lambda_memory_size" {
  description = "Lambda function memory size in MB"
  type        = number
  default     = 512

  validation {
    condition     = var.lambda_memory_size >= 128 && var.lambda_memory_size <= 10240
    error_message = "Lambda memory size must be between 128 and 10240 MB."
  }
}

variable "lambda_log_retention_days" {
  description = "CloudWatch log retention in days for Lambda (0 = never expire)"
  type        = number
  default     = 0

  validation {
    condition = var.lambda_log_retention_days == 0 || contains([
      1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180,
      365, 400, 545, 731, 1827, 3653
    ], var.lambda_log_retention_days)
    error_message = "Log retention days must be 0 (never expire) or a valid CloudWatch Logs retention value."
  }
}

variable "schedule_group_name" {
  description = "EventBridge Scheduler schedule group name"
  type        = string
  default     = "cierge"
}

# Data Sources

data "aws_caller_identity" "current" {}

# KMS Key

resource "aws_kms_key" "cierge" {
  description             = "Cierge encryption key for tokens and callback secrets"
  deletion_window_in_days = var.kms_deletion_window
  enable_key_rotation     = true

  tags = {
    Name = "cierge-${var.environment}"
  }
}

resource "aws_kms_alias" "cierge" {
  name          = "alias/cierge-${var.environment}"
  target_key_id = aws_kms_key.cierge.key_id
}

resource "aws_kms_key_policy" "cierge" {
  key_id = aws_kms_key.cierge.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowRootReadAndDelete"
        Effect = "Allow"
        # Admins can read key state and delete key if all principals are gone
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GetKeyPolicy",
          "kms:PutKeyPolicy",
          "kms:ListKeyPolicies",
          "kms:GetKeyRotationStatus",
          "kms:ListResourceTags",
          "kms:ListAliases",
          "kms:ScheduleKeyDeletion"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowLambdaDecrypt"
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_role.lambda_execution.arn
        }
        Action   = "kms:Decrypt"
        Resource = "*"
      },
      {
        Sid    = "AllowServerEncryptDecrypt"
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_user.cierge_server.arn
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:Encrypt",
          "kms:Decrypt"
        ]
        Resource = "*"
      },
      {
        # Scheduler execution role needs kms:Decrypt to decrypt the schedule payload at run time
        Sid    = "AllowSchedulerExecutionDecrypt"
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_role.scheduler_execution.arn
        }
        Action   = "kms:Decrypt"
        Resource = "*"
      }
    ]
  })
}

# EventBridge Scheduler Group

resource "aws_scheduler_schedule_group" "cierge" {
  name = var.schedule_group_name

  tags = {
    Name = "cierge-${var.environment}"
  }
}

# IAM User for Cierge Server

resource "aws_iam_user" "cierge_server" {
  name = "cierge-server-${var.environment}"
  path = "/cierge/"

  tags = {
    Name = "cierge-server-${var.environment}"
  }
}

resource "aws_iam_access_key" "cierge_server" {
  user = aws_iam_user.cierge_server.name
}

resource "aws_iam_user_policy" "cierge_server" {
  name = "cierge-server-permissions"
  user = aws_iam_user.cierge_server.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "KMSEncryptDecrypt"
        Effect = "Allow"
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:Encrypt",
          "kms:Decrypt"
        ]
        # Scoped to only the Cierge KMS key
        Resource = aws_kms_key.cierge.arn
      },
      {
        Sid    = "EventBridgeScheduler"
        Effect = "Allow"
        Action = [
          "scheduler:CreateSchedule",
          "scheduler:UpdateSchedule",
          "scheduler:DeleteSchedule",
          "scheduler:GetSchedule"
        ]
        # Scoped to only cierge-job-* schedules in the cierge group
        Resource = "arn:aws:scheduler:${var.aws_region}:${data.aws_caller_identity.current.account_id}:schedule/${var.schedule_group_name}/cierge-job-*"
      },
      {
        Sid    = "PassSchedulerRole"
        Effect = "Allow"
        Action = "iam:PassRole"
        # Scoped to only the scheduler execution role
        Resource = aws_iam_role.scheduler_execution.arn
        Condition = {
          StringEquals = {
            "iam:PassedToService" = "scheduler.amazonaws.com"
          }
        }
      }
    ]
  })
}

# IAM Role for EventBridge Scheduler

resource "aws_iam_role" "scheduler_execution" {
  name = "cierge-scheduler-execution-${var.environment}"
  path = "/cierge/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "scheduler.amazonaws.com"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })

  tags = {
    Name = "cierge-scheduler-execution-${var.environment}"
  }
}

resource "aws_iam_role_policy" "scheduler_execution" {
  name = "invoke-lambda"
  role = aws_iam_role.scheduler_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "lambda:InvokeFunction"
        # Scoped to only the reservation handler Lambda
        Resource = aws_lambda_function.reservation_handler.arn
      },
      {
        Sid    = "KMSDecrypt"
        Effect = "Allow"
        Action = "kms:Decrypt"
        # Needed to decrypt the schedule payload encrypted with the customer-managed key
        Resource = aws_kms_key.cierge.arn
      }
    ]
  })
}

# Lambda Function

resource "aws_cloudwatch_log_group" "lambda" {
  name = "/aws/lambda/cierge-reservation-handler-${var.environment}"
  # 0 = never expire, otherwise specific retention period
  retention_in_days = var.lambda_log_retention_days == 0 ? null : var.lambda_log_retention_days

  tags = {
    Name = "cierge-lambda-logs-${var.environment}"
  }
}

resource "aws_iam_role" "lambda_execution" {
  name = "cierge-lambda-execution-${var.environment}"
  path = "/cierge/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Name = "cierge-lambda-execution-${var.environment}"
  }
}

resource "aws_iam_role_policy" "lambda_execution" {
  name = "lambda-execution"
  role = aws_iam_role.lambda_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "CloudWatchLogs"
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        # Scoped to only this Lambda's log group
        Resource = "${aws_cloudwatch_log_group.lambda.arn}:*"
      },
      {
        Sid    = "KMSDecrypt"
        Effect = "Allow"
        Action = "kms:Decrypt"
        # Scoped to only the Cierge KMS key
        Resource = aws_kms_key.cierge.arn
      }
    ]
  })
}

# Trigger rebuild when any .go file or dependencies change in lambda/
resource "null_resource" "build_lambda" {
  triggers = {
    # Hash of all Go source files in lambda directory
    go_sources = sha256(join("", [
      for f in fileset("${path.module}/../lambda", "*.go") :
      filesha256("${path.module}/../lambda/${f}")
    ]))
    # Hash of dependency files
    go_mod       = filesha256("${path.module}/../lambda/go.mod")
    go_sum       = filesha256("${path.module}/../lambda/go.sum")
    build_script = filesha256("${path.module}/../lambda/build.sh")
  }

  provisioner "local-exec" {
    command     = "./build.sh"
    working_dir = "${path.module}/../lambda"
  }
}

data "archive_file" "lambda" {
  type        = "zip"
  source_dir  = "${path.module}/../lambda"
  output_path = "${path.module}/build/lambda.zip"
  # Only include the built bootstrap binary, exclude source files
  excludes = [
    "*.go",
    "go.mod",
    "go.sum",
    "*.md",
    "*.sh"
  ]

  depends_on = [null_resource.build_lambda]
}

resource "aws_lambda_function" "reservation_handler" {
  filename         = data.archive_file.lambda.output_path
  function_name    = "cierge-reservation-handler-${var.environment}"
  role             = aws_iam_role.lambda_execution.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architectures    = ["arm64"]
  timeout          = var.lambda_timeout
  memory_size      = var.lambda_memory_size
  source_code_hash = data.archive_file.lambda.output_base64sha256

  environment {
    variables = {
      ENVIRONMENT = var.environment
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.lambda,
    null_resource.build_lambda
  ]

  tags = {
    Name = "cierge-reservation-handler-${var.environment}"
  }
}

# Outputs

output "kms_key_arn" {
  description = "ARN of the KMS key (use for kms_key_id in config.json)"
  value       = aws_kms_key.cierge.arn
}

output "kms_key_id" {
  description = "ID of the KMS key"
  value       = aws_kms_key.cierge.key_id
}

output "lambda_arn" {
  description = "ARN of the reservation handler Lambda (use for lambda_arn in config.json)"
  value       = aws_lambda_function.reservation_handler.arn
}

output "scheduler_role_arn" {
  description = "ARN of the EventBridge Scheduler execution role (use for scheduler_role_arn in config.json)"
  value       = aws_iam_role.scheduler_execution.arn
}

output "schedule_group_name" {
  description = "EventBridge Scheduler schedule group name"
  value       = aws_scheduler_schedule_group.cierge.name
}

output "server_access_key_id" {
  description = "AWS Access Key ID for the Cierge server (use for access_key_id in config.json or AWS_ACCESS_KEY_ID env var)"
  value       = aws_iam_access_key.cierge_server.id
  sensitive   = true
}

output "server_secret_access_key" {
  description = "AWS Secret Access Key for the Cierge server (use for secret_access_key in config.json or AWS_SECRET_ACCESS_KEY env var)"
  value       = aws_iam_access_key.cierge_server.secret
  sensitive   = true
}

output "region" {
  description = "AWS region (use for region in config.json)"
  value       = var.aws_region
}

output "config_summary" {
  description = "Summary of values for config.json cloud section"
  value = {
    provider            = "aws"
    region              = var.aws_region
    kms_key_id          = aws_kms_key.cierge.arn
    lambda_arn          = aws_lambda_function.reservation_handler.arn
    scheduler_role_arn  = aws_iam_role.scheduler_execution.arn
    schedule_group_name = aws_scheduler_schedule_group.cierge.name
    credentials_note    = "Run 'terraform output -raw server_access_key_id' and 'terraform output -raw server_secret_access_key' for credentials (or use AWS env vars)"
  }
}
