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
	"github.com/charmbracelet/lipgloss"
	"github.com/daylamtayari/cierge/api"
	"github.com/daylamtayari/cierge/resy"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	restaurantPlatformId    string
	jobReservationDateInput string
	jobReservationDate      *time.Time
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

			if cmd.Flags().Changed("platform") {
				jobPlatform = strings.ToLower(jobPlatform)
				if jobPlatform != "resy" && jobPlatform != "opentable" {
					logger.Fatal().Msgf("Invalid platform %q specified - only 'resy' and 'opentable' are supported platforms", jobPlatform)
				}
			} else {
				err := huh.NewSelect[string]().
					Title("Select reservation platform:").
					Options(
						huh.NewOption("Resy", "resy"),
						huh.NewOption("OpenTable", "opentable"),
					).
					Value(&jobPlatform).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for reservation platform")
				}
			}

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

			if cmd.Flags().Changed("size") && jobPartySize <= 0 {
				logger.Error().Msg("Party size must be greater than 0")
				jobPartySize = 0
			}
			if jobPartySize == 0 {
				var partySize string
				err := huh.NewInput().
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
					}).
					Run()
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
					jobReservationDate = &reservationDate
				}
			}
			if jobReservationDate == nil {
				var dateInput string
				err := huh.NewInput().
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
					}).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for reservation date")
				}

				parsedDate, _ := time.Parse("02-01-2006", dateInput)
				jobReservationDate = &parsedDate
			}

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
				for _, dc := range dropConfigs {
					dropDate := jobReservationDate.Add(-time.Duration(dc.DaysInAdvance) * 24 * time.Hour)
					label := fmt.Sprintf("%*d %s  %*d days in advance (%s) at %s %s", confWidth, dc.Confidence, upArrow, daysWidth, dc.DaysInAdvance, dropDate.Format("02 Jan"), dc.DropTime, tzAbbr)
					options = append(options, huh.NewOption(label, dc.ID.String()))
				}
				options = append(options, huh.NewOption("Create new drop configuration", "new"))

				var selectedDropConfig string
				err := huh.NewSelect[string]().
					Title("Select drop configuration:").
					Options(options...).
					Value(&selectedDropConfig).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for drop configuration")
				}
				if selectedDropConfig != "new" {
					parsedId, _ := uuid.Parse(selectedDropConfig)
					dropConfig = &parsedId
				}
			}
			if dropConfig == nil {
				var daysInAdvanceInput string
				err := huh.NewInput().
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
					}).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for days in advance")
				}

				var dropTimeInput string
				err = huh.NewInput().
					Title("Drop time (HH:mm):").
					Placeholder("09:00").
					Value(&dropTimeInput).
					Validate(func(s string) error {
						if _, err := time.Parse("15:04", s); err != nil {
							return errors.New("invalid time format - use HH:mm")
						}
						return nil
					}).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for drop time")
				}

				daysInAdvance, _ := strconv.ParseInt(daysInAdvanceInput, 10, 16)
				dropTimeParsed, _ := time.Parse("15:04", dropTimeInput)
				dropDate := jobReservationDate.Add(-time.Duration(daysInAdvance) * 24 * time.Hour)
				loc, err := time.LoadLocation(restaurant.Timezone)
				if err != nil {
					loc = time.UTC
				}
				expectedDrop := time.Date(dropDate.Year(), dropDate.Month(), dropDate.Day(), dropTimeParsed.Hour(), dropTimeParsed.Minute(), 0, 0, loc)

				var confirmed bool
				err = huh.NewConfirm().
					Title("Confirm drop configuration").
					Description(fmt.Sprintf("%d days in advance at %s — expected drop: %s", daysInAdvance, dropTimeInput, expectedDrop.Format("02 Jan at 15:04 MST"))).
					Value(&confirmed).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to confirm drop configuration")
				}
				if !confirmed {
					logger.Fatal().Msg("Drop configuration creation cancelled")
				}

				newDropConfig, err := client.CreateDropConfig(restaurant.ID, int16(daysInAdvance), dropTimeInput)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to create drop configuration")
				}
				dropConfig = &newDropConfig.ID
			}
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

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

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
	ti := textinput.New()
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
	selected      map[string]bool
	cursor        int
	filterInput   textinput.Model
	filtering     bool
	confirmed     bool
	quitting      bool
	err           error
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
			if len(m.selected) == 0 {
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
				if m.selected[slot] {
					delete(m.selected, slot)
				} else {
					m.selected[slot] = true
				}
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

	n := len(m.filteredSlots)
	if n >= timeSlotViewHeight {
		// Carousel: cursor is always pinned at the centre of the visible window.
		// Each navigation step shifts the entire window by one, wrapping seamlessly.
		half := timeSlotViewHeight / 2
		startIdx := (m.cursor - half + n) % n
		for i := range timeSlotViewHeight {
			slotIdx := (startIdx + i) % n
			slot := m.filteredSlots[slotIdx]
			isCursor := i == half
			check := " "
			if m.selected[slot] {
				check = "✗"
			}
			cur := " "
			if isCursor {
				cur = ">"
			}
			line := fmt.Sprintf("%s [%s] %s", cur, check, slot)
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
			check := " "
			if m.selected[slot] {
				check = "✗"
			}
			cur := " "
			if isCursor {
				cur = ">"
			}
			line := fmt.Sprintf("%s [%s] %s", cur, check, slot)
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
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(m.err.Error()))
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render(fmt.Sprintf("%d selected • ↑/↓: navigate • space/x: toggle • f//: search • enter: confirm • esc: cancel", len(m.selected))))

	return b.String()
}

// runTimeSlotPicker presents an interactive multi-select time slot picker.
// Slots are ordered starting from 18:00 and wrap around to 17:45.
// Returns selected slots in chronological order.
func runTimeSlotPicker() ([]string, error) {
	allSlots := make([]string, 96)
	for i := range 96 {
		slotIndex := (72 + i) % 96
		hour := slotIndex / 4
		minute := (slotIndex % 4) * 15
		allSlots[i] = fmt.Sprintf("%02d:%02d", hour, minute)
	}

	fi := textinput.New()
	fi.Placeholder = "type to filter..."
	fi.CharLimit = 5

	m := timeSlotModel{
		allSlots:      allSlots,
		filteredSlots: allSlots,
		selected:      make(map[string]bool),
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

	// Return slots in chronological order, not selection order
	var selected []string
	for _, slot := range result.allSlots {
		if result.selected[slot] {
			selected = append(selected, slot)
		}
	}
	return selected, nil
}
