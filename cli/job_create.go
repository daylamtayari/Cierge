package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/daylamtayari/cierge/api"
	"github.com/daylamtayari/cierge/resy"
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	restaurantPlatformId    string
	jobReservationDateInput string
	jobReservationDate      *string
	jobPartySize            int16
	jobPlatform             string
	jobTimeSlotsInput       []string
	jobTimeSlots            []string
	jobDropConfigId         string

	jobCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new reservation job",
		Run: func(cmd *cobra.Command, args []string) {
			client := newClient()
			resyClient := resy.NewClient(nil, resy.Tokens{ApiKey: resy.DefaultApiKey}, "")

			// Platform selection
			if cmd.Flags().Changed("platform") {
				jobPlatform = strings.ToLower(jobPlatform)
				if jobPlatform != "resy" && jobPlatform != "opentable" {
					logger.Fatal().Msgf("Invalid platform %q specified - only 'resy' and 'opentable' are supported platforms", jobPlatform)
				}
			} else {
				err := runHuh(huh.NewSelect[string]().
					Title("Select reservation platform:").
					Options(
						huh.NewOption("Resy", "resy"),
						huh.NewOption("OpenTable", "opentable"),
					).
					Value(&jobPlatform))
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for reservation platform")
				}
			}
			platformTokens, err := client.GetPlatformTokens(&jobPlatform)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to retrieve platform tokens")
			} else if len(platformTokens) == 0 {
				logger.Fatal().Msgf("No platform tokens registered for %q, please add tokens first using the 'token add' command", jobPlatform)
			}

			// Restaurant selection
			var restaurant *api.Restaurant
			if cmd.Flags().Changed("restaurant") {
				switch jobPlatform {
				case "resy":
					resyPlatformId, err := strconv.Atoi(restaurantPlatformId)
					if err != nil {
						logger.Error().Err(err).Msg("Resy restaurant platform ID must be numerical")
					}

					_, err = resyClient.GetVenue(resyPlatformId)
					if err != nil && errors.Is(err, resy.ErrNotFound) {
						logger.Error().Err(err).Msg("Invalid Resy restaurant ID")
					} else if err != nil {
						logger.Error().Err(err).Msg("Failed to fetch Resy restaurant")
					} else {
						res, err := client.GetRestaurantByPlatform(jobPlatform, restaurantPlatformId)
						if err != nil {
							logger.Error().Err(err).Msg("Failed to get restaurant")
						} else {
							restaurant = &res
						}
					}

				case "opentable":
					// TODO: Implement opentable
				}
			}
			if restaurant == nil {
				switch jobPlatform {
				case "resy":
					venueId, err := runResyVenueSearch(resyClient)
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to search for venue")
					}
					restaurantPlatformId = strconv.Itoa(venueId)
				case "opentable":
					// TODO: Implement opentable search
					logger.Fatal().Msg("OpenTable search not yet implemented")
				}

				res, err := client.GetRestaurantByPlatform(jobPlatform, restaurantPlatformId)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to get restaurant")
				}
				restaurant = &res
			}

			// Party size selection
			if cmd.Flags().Changed("size") && jobPartySize <= 0 {
				logger.Error().Msg("Party size must be greater than 0")
				jobPartySize = 0
			}
			if jobPartySize == 0 {
				var partySize string
				err := runHuh(huh.NewInput().
					Title("Enter party size:").
					Value(&partySize).
					Validate(func(s string) error {
						val, err := strconv.ParseInt(s, 10, 16)
						if err != nil {
							return errors.New("party size must be a valid number")
						}
						if val <= 0 {
							return errors.New("party size must be greater than 0")
						}
						return nil
					}))
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for party size")
				}

				val, _ := strconv.ParseInt(partySize, 10, 16)
				jobPartySize = int16(val)
			}

			if cmd.Flags().Changed("date") {
				reservationDate, err := time.Parse("02-01-2006", jobReservationDateInput)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to parse specified reservation date")
				} else {
					reservationDateFormatted := reservationDate.Format("2006-01-02")
					jobReservationDate = &reservationDateFormatted
				}
			}
			if jobReservationDate == nil {
				var dateInput string
				err := runHuh(huh.NewInput().
					Title("Enter reservation date (DD-MM-YYYY):").
					Placeholder("01-12-2026").
					Value(&dateInput).
					Validate(func(s string) error {
						if s == "" {
							return errors.New("reservation date is required")
						}
						parsedDate, err := time.Parse("02-01-2006", s)
						if err != nil {
							return errors.New("invalid date format - use DD-MM-YYYY")
						}
						// Validate that the date is not in the past
						if parsedDate.Before(time.Now().Truncate(24 * time.Hour)) {
							return errors.New("reservation date cannot be in the past")
						}
						return nil
					}))
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for reservation date")
				}

				parsedDate, _ := time.Parse("02-01-2006", dateInput)
				parsedDateFormatted := parsedDate.Format("2006-01-02")
				jobReservationDate = &parsedDateFormatted
			}

			// Timeslot selection
			if len(jobTimeSlotsInput) > 0 {
				for _, tsInput := range jobTimeSlotsInput {
					if _, err := time.Parse("15:04", tsInput); err == nil {
						jobTimeSlots = append(jobTimeSlots, tsInput)
					} else {
						logger.Error().Err(err).Msg("time slot is in invalid format")
					}
				}
			}
			if len(jobTimeSlots) == 0 {
				var err error
				jobTimeSlots, err = runTimeSlotPicker()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for time slots")
				}
			}

			// Drop config selection
			var dropConfig *uuid.UUID
			dropConfigs, err := client.GetDropConfigs(restaurant.ID)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to retrieve drop configurations")
			}
			if cmd.Flags().Changed("drop-config") {
				parsedId, err := uuid.Parse(jobDropConfigId)
				if err != nil {
					logger.Error().Err(err).Msgf("Invalid drop config ID %q - must be a valid UUID", jobDropConfigId)
				} else {
					for _, dc := range dropConfigs {
						if dc.ID == parsedId {
							dropConfig = &parsedId
							break
						}
					}
					if dropConfig == nil {
						logger.Error().Msgf("Drop config ID %q was not found for the restaurant", jobDropConfigId)
					}
				}
			}
			if len(dropConfigs) > 0 {
				options := make([]huh.Option[string], 0, len(dropConfigs)+1)
				tzAbbr := restaurant.Timezone
				if loc, err := time.LoadLocation(restaurant.Timezone); err == nil {
					tzAbbr = time.Now().In(loc).Format("MST")
				}
				maxConf, maxDays := 0, 0
				for _, dc := range dropConfigs {
					if c := int(dc.Confidence); c > maxConf {
						maxConf = c
					}
					if d := int(dc.DaysInAdvance); d > maxDays {
						maxDays = d
					}
				}
				confWidth := len(fmt.Sprintf("%d", maxConf))
				daysWidth := len(fmt.Sprintf("%d", maxDays))
				reservationDate, _ := time.Parse("2006-01-02", *jobReservationDate)
				for _, dc := range dropConfigs {
					dropDate := reservationDate.Add(-time.Duration(dc.DaysInAdvance) * 24 * time.Hour)
					label := fmt.Sprintf("%*d %s  %*d days in advance (%s) at %s %s", confWidth, dc.Confidence, upArrow, daysWidth, dc.DaysInAdvance, dropDate.Format("02 Jan"), dc.DropTime, tzAbbr)
					options = append(options, huh.NewOption(label, dc.ID.String()))
				}
				options = append(options, huh.NewOption("Create new drop configuration", "new"))

				var selectedDropConfig string
				err := runHuh(huh.NewSelect[string]().
					Title("Select drop configuration:").
					Options(options...).
					Value(&selectedDropConfig))
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for drop configuration")
				}
				if selectedDropConfig != "new" {
					parsedId, _ := uuid.Parse(selectedDropConfig)
					dropConfig = &parsedId
				}
			}
			if dropConfig == nil {
				var daysInAdvanceInput, dropTimeInput string
				for {
					err := runHuh(huh.NewInput().
						Title("Days in advance:").
						Description("How many days before the reservation date should the drop be attempted?").
						Value(&daysInAdvanceInput).
						Validate(func(s string) error {
							val, err := strconv.ParseInt(s, 10, 16)
							if err != nil {
								return errors.New("days in advance must be a valid number")
							}
							if val <= 0 {
								return errors.New("days in advance must be greater than 0")
							}
							return nil
						}))
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to prompt user for days in advance")
					}

					err = runHuh(huh.NewInput().
						Title("Drop time (HH:mm):").
						Placeholder("09:00").
						Value(&dropTimeInput).
						Validate(func(s string) error {
							if _, err := time.Parse("15:04", s); err != nil {
								return errors.New("invalid time format - use HH:mm")
							}
							return nil
						}))
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to prompt user for drop time")
					}

					reservationDate, _ := time.Parse("2006-01-02", *jobReservationDate)
					daysInAdvance, _ := strconv.ParseInt(daysInAdvanceInput, 10, 16)
					dropTimeParsed, _ := time.Parse("15:04", dropTimeInput)
					dropDate := reservationDate.Add(-time.Duration(daysInAdvance) * 24 * time.Hour)
					loc, err := time.LoadLocation(restaurant.Timezone)
					if err != nil {
						loc = time.UTC
					}
					expectedDrop := time.Date(dropDate.Year(), dropDate.Month(), dropDate.Day(), dropTimeParsed.Hour(), dropTimeParsed.Minute(), 0, 0, loc)

					var confirmed bool
					err = runHuh(huh.NewConfirm().
						Title("Confirm drop configuration").
						Description(fmt.Sprintf("%d days in advance at %s — expected drop: %s", daysInAdvance, dropTimeInput, expectedDrop.Format("02 Jan at 15:04 MST"))).
						Value(&confirmed))
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to confirm drop configuration")
					}
					if confirmed {
						break
					}
				}

				daysInAdvance, _ := strconv.ParseInt(daysInAdvanceInput, 10, 16)
				newDropConfig, err := client.CreateDropConfig(restaurant.ID, int16(daysInAdvance), dropTimeInput)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to create drop configuration")
				}
				dropConfig = &newDropConfig.ID
			}

			// Job creation
			job, err := client.CreateJob(api.JobCreationRequest{
				RestaurantID:    restaurant.ID,
				ReservationDate: *jobReservationDate,
				PartySize:       jobPartySize,
				PreferredTimes:  jobTimeSlots,
				DropConfigID:    *dropConfig,
			})
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to create job")
			}

			reservationDate, _ := time.Parse("2006-01-02", job.ReservationDate)
			jt := table.NewWriter()
			jt.SetStyle(table.StyleLight)
			jt.Style().Options.DrawBorder = false
			jt.Style().Options.SeparateColumns = false
			jt.AppendRows([]table.Row{
				{"ID", job.ID},
				{"Restaurant", restaurant.Name},
				{"Platform", job.Platform},
				{"Scheduled At", job.ScheduledAt.Local().Format("02 January at 15:04 MST")},
				{"Reservation Date", reservationDate.Format("02 January 2006")},
				{"Party Size", job.PartySize},
				{"Preferred Times", job.PreferredTimes},
			})
			fmt.Print(jt.Render() + "\n")
		},
	}
)

func initJobCreateCmd() *cobra.Command {
	jobCreateCmd.Flags().StringVar(&jobPlatform, "platform", "", "Platform to book with")
	jobCreateCmd.Flags().Int16Var(&jobPartySize, "size", 0, "Size of the party")
	jobCreateCmd.Flags().StringVar(&jobReservationDateInput, "date", "", "Date for the reservation - format: DD-MM-YYYY")
	jobCreateCmd.Flags().StringVar(&restaurantPlatformId, "restaurant", "", "ID of the restaurant for the respective platform")
	jobCreateCmd.Flags().StringSliceVar(&jobTimeSlotsInput, "slots", nil, "Time slots for the reservation - format: HH:mm")
	jobCreateCmd.Flags().StringVar(&jobDropConfigId, "drop-config", "", "ID of the drop configuration to use")
	return jobCreateCmd
}

// resySearchModel implements a real-time venue search with debouncing
type resySearchModel struct {
	client        *resy.Client
	textInput     textinput.Model
	venues        []resy.Venue
	cursor        int
	selectedVenue *resy.Venue
	lastQuery     string
	searchPending bool // Waiting for debounce timer
	searching     bool // Actively performing API call
	err           error
	quitting      bool
}

type searchResultMsg struct {
	venues []resy.Venue
	err    error
}

type triggerSearchMsg struct{}

const searchDebounceMs = 300

func (m resySearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m resySearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.cursor < len(m.venues) && len(m.venues) > 0 {
				m.selectedVenue = &m.venues[m.cursor]
				m.quitting = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.venues)-1 {
				m.cursor++
			}
		}

	case searchResultMsg:
		m.searching = false
		m.searchPending = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.venues = msg.venues
			m.cursor = 0 // Reset cursor to top when results update
		}

	case triggerSearchMsg:
		currentQuery := m.textInput.Value()
		// Only search if query changed and is non-empty
		if currentQuery != m.lastQuery {
			if currentQuery == "" {
				// Clear results when query is empty
				m.lastQuery = ""
				m.venues = []resy.Venue{}
				m.cursor = 0
				m.err = nil
			} else {
				m.lastQuery = currentQuery
				m.searching = true
				m.searchPending = false
				return m, m.performSearch(currentQuery)
			}
		}
		m.searchPending = false
	}

	oldValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)
	newValue := m.textInput.Value()

	// Schedule a search if text changed and no search is already pending
	if oldValue != newValue && !m.searchPending {
		m.searchPending = true
		return m, tea.Tick(searchDebounceMs*time.Millisecond, func(t time.Time) tea.Msg {
			return triggerSearchMsg{}
		})
	}

	return m, cmd
}

func (m resySearchModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Search for a restaurant"))
	b.WriteString("\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.err != nil {
		fmt.Fprintf(&b, "Error: %v\n\n", m.err)
	}

	if m.searching {
		b.WriteString("Searching...\n\n")
	} else if len(m.venues) > 0 {
		b.WriteString("Results:\n")
		for i, venue := range m.venues {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			if m.cursor == i {
				line := fmt.Sprintf("%s %s - %s, %s", cursor, venue.Name, venue.Locality, venue.Region)
				b.WriteString(selectedStyle.Render(line))
			} else {
				fmt.Fprintf(&b, "%s %s - %s, %s", cursor, venue.Name, venue.Locality, venue.Region)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	} else if m.textInput.Value() != "" && !m.searching {
		b.WriteString("No results found.\n\n")
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: cancel"))

	return b.String()
}

func (m resySearchModel) performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		venues, err := m.client.SearchVenue(query, nil)
		return searchResultMsg{
			venues: venues,
			err:    err,
		}
	}
}

// runResyVenueSearch provides an interactive real-time search interface for Resy venues.
// It returns the venue ID of the selected restaurant or an error if cancelled or failed.
func runResyVenueSearch(client *resy.Client) (int, error) {
	ti := styledTextInput()
	ti.Placeholder = "Type to search..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	m := resySearchModel{
		client:    client,
		textInput: ti,
		venues:    []resy.Venue{},
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return 0, fmt.Errorf("error running search: %w", err)
	}

	result := finalModel.(resySearchModel)
	if result.selectedVenue == nil {
		return 0, fmt.Errorf("no venue selected")
	}

	return result.selectedVenue.Id.Resy, nil
}

const timeSlotViewHeight = 10

type timeSlotModel struct {
	allSlots      []string
	filteredSlots []string
	prioritySlots []string // ordered by priority; index 0 = priority 1
	cursor        int
	filterInput   textinput.Model
	filtering     bool
	confirmed     bool
	quitting      bool
	err           error
}

// slotPriority returns the 1-based priority of slot, or 0 if unselected.
func (m timeSlotModel) slotPriority(slot string) int {
	for i, s := range m.prioritySlots {
		if s == slot {
			return i + 1
		}
	}
	return 0
}

func (m timeSlotModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m timeSlotModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = nil
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.filtering {
				m.filtering = false
				m.filterInput.Blur()
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.filtering {
				m.filtering = false
				m.filterInput.Blur()
				return m, nil
			}
			if len(m.prioritySlots) == 0 {
				m.err = errors.New("select at least one time slot")
				return m, nil
			}
			m.confirmed = true
			return m, tea.Quit

		case "f", "/":
			if !m.filtering {
				m.filtering = true
				return m, m.filterInput.Focus()
			}

		case "up", "k":
			if !m.filtering && len(m.filteredSlots) > 0 {
				m.cursor = (m.cursor - 1 + len(m.filteredSlots)) % len(m.filteredSlots)
			}

		case "down", "j":
			if !m.filtering && len(m.filteredSlots) > 0 {
				m.cursor = (m.cursor + 1) % len(m.filteredSlots)
			}

		case " ", "x":
			if !m.filtering && len(m.filteredSlots) > 0 {
				slot := m.filteredSlots[m.cursor]
				if p := m.slotPriority(slot); p > 0 {
					m.prioritySlots = tsRemoveAt(m.prioritySlots, p-1)
				} else {
					m.prioritySlots = append(m.prioritySlots, slot)
				}
			}

		case "+":
			// Increase priority (lower number, move toward front of list)
			if !m.filtering && len(m.filteredSlots) > 0 {
				slot := m.filteredSlots[m.cursor]
				if p := m.slotPriority(slot); p > 1 {
					m.prioritySlots = tsSwapAt(m.prioritySlots, p-2, p-1)
				}
			}

		case "-":
			// Decrease priority (higher number, move toward end of list)
			if !m.filtering && len(m.filteredSlots) > 0 {
				slot := m.filteredSlots[m.cursor]
				if p := m.slotPriority(slot); p > 0 && p < len(m.prioritySlots) {
					m.prioritySlots = tsSwapAt(m.prioritySlots, p-1, p)
				}
			}

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if !m.filtering && len(m.filteredSlots) > 0 {
				n, _ := strconv.Atoi(msg.String())
				slot := m.filteredSlots[m.cursor]
				// Remove from current position if already selected
				if p := m.slotPriority(slot); p > 0 {
					m.prioritySlots = tsRemoveAt(m.prioritySlots, p-1)
				}
				// Insert at desired position, clamped to valid range
				insertPos := n - 1
				if insertPos > len(m.prioritySlots) {
					insertPos = len(m.prioritySlots)
				}
				m.prioritySlots = tsInsertAt(m.prioritySlots, insertPos, slot)
			}
		}
	}

	if m.filtering {
		prevValue := m.filterInput.Value()
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		if m.filterInput.Value() != prevValue {
			query := strings.ToLower(m.filterInput.Value())
			if query == "" {
				m.filteredSlots = m.allSlots
			} else {
				m.filteredSlots = nil
				for _, slot := range m.allSlots {
					if strings.Contains(slot, query) {
						m.filteredSlots = append(m.filteredSlots, slot)
					}
				}
			}
			m.cursor = 0
		}
		return m, cmd
	}

	return m, nil
}

func (m timeSlotModel) View() string {
	if m.quitting || m.confirmed {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Select desired time slots:"))
	b.WriteString("\n")
	if m.filtering {
		b.WriteString("Search: ")
		b.WriteString(m.filterInput.View())
		b.WriteString("\n")
	} else if m.filterInput.Value() != "" {
		b.WriteString(helpStyle.Render("Filter: " + m.filterInput.Value()))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Keep bracket width consistent as selections grow to avoid layout jitter
	numWidth := len(fmt.Sprintf("%d", max(len(m.prioritySlots), 1)))
	renderSlot := func(slot string, isCursor bool) string {
		cur := " "
		if isCursor {
			cur = ">"
		}
		var check string
		if p := m.slotPriority(slot); p > 0 {
			check = fmt.Sprintf("%*d", numWidth, p)
		} else {
			check = strings.Repeat(" ", numWidth)
		}
		return fmt.Sprintf("%s [%s] %s", cur, check, slot)
	}

	n := len(m.filteredSlots)
	if n >= timeSlotViewHeight {
		// Carousel: cursor pinned at centre, window shifts each step.
		half := timeSlotViewHeight / 2
		startIdx := (m.cursor - half + n) % n
		for i := range timeSlotViewHeight {
			slotIdx := (startIdx + i) % n
			isCursor := i == half
			line := renderSlot(m.filteredSlots[slotIdx], isCursor)
			if isCursor {
				b.WriteString(selectedStyle.Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
	} else {
		for i, slot := range m.filteredSlots {
			isCursor := i == m.cursor
			line := renderSlot(slot, isCursor)
			if isCursor {
				b.WriteString(selectedStyle.Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render(fmt.Sprintf(
		"%d selected • ↑/k / ↓/j: navigate • space/x: toggle • 1-9: set priority • +/-: adjust priority • f / /: search • enter: confirm • esc: cancel",
		len(m.prioritySlots),
	)))

	return b.String()
}

func tsRemoveAt(s []string, i int) []string {
	out := make([]string, 0, len(s)-1)
	out = append(out, s[:i]...)
	return append(out, s[i+1:]...)
}

func tsInsertAt(s []string, i int, v string) []string {
	out := make([]string, 0, len(s)+1)
	out = append(out, s[:i]...)
	out = append(out, v)
	return append(out, s[i:]...)
}

func tsSwapAt(s []string, i, j int) []string {
	out := make([]string, len(s))
	copy(out, s)
	out[i], out[j] = out[j], out[i]
	return out
}

// runTimeSlotPicker presents an interactive multi-select time slot picker.
// Slots are ordered starting from 18:00 and wrap around to 17:45.
// Returns selected slots in priority order.
func runTimeSlotPicker() ([]string, error) {
	allSlots := make([]string, 96)
	for i := range 96 {
		slotIndex := (72 + i) % 96
		hour := slotIndex / 4
		minute := (slotIndex % 4) * 15
		allSlots[i] = fmt.Sprintf("%02d:%02d", hour, minute)
	}

	fi := styledTextInput()
	fi.Placeholder = "type to filter..."
	fi.CharLimit = 5

	m := timeSlotModel{
		allSlots:      allSlots,
		filteredSlots: allSlots,
		filterInput:   fi,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running time slot picker: %w", err)
	}

	result := finalModel.(timeSlotModel)
	if !result.confirmed {
		return nil, fmt.Errorf("no time slots selected")
	}

	return result.prioritySlots, nil
}
