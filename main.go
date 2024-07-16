package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nelsonnyan2001/jwt-helper/aws"
)

const path = "/.swing-jwt-helper.csv"

type model struct {
	focusScreen int
	focusIndex  int
	inputs      []textinput.Model
}

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, 3),
	}

	for i := range m.inputs {
		t := textinput.New()
		switch i {
		case 0:
			t.Placeholder = "1e9lq2b44vvvv3v89v7vvv4vvv"
			t.Focus()
		case 1:
			t.Placeholder = "sub@gmail.com"
		case 2:
			t.Placeholder = "********"
		}
		m.inputs[i] = t
	}
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("Swing JWT Helper"), textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.focusScreen == 1 {
				if m.focusIndex == 2 {
					m.inputs[m.focusIndex].Blur()
					m.focusScreen++
					m.focusIndex = 0
				} else {
					m.inputs[m.focusIndex].Blur()
					m.inputs[m.focusIndex+1].Focus()
					m.focusIndex++
				}
			}
			if m.focusScreen == 0 {
				m.focusScreen++
			}
		case "up":
			if m.focusScreen == 1 {
				if m.focusIndex != 0 {
					m.inputs[m.focusIndex].Blur()
					m.inputs[m.focusIndex-1].Focus()
					m.focusIndex--
				}
			}
		case "n":
			if m.focusScreen == 2 {
				m.focusScreen = 1
				m.inputs[0].SetValue("")
				m.inputs[1].SetValue("")
				m.inputs[2].SetValue("")
				m.inputs[2].Blur()
				m.inputs[0].Focus()
			}
		case "y":
			if m.focusScreen == 2 {
				createFile(getFilePath(), m.inputs[0].Value(), m.inputs[1].Value(), m.inputs[2].Value())
				m.focusScreen++
			}
		}
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)

	}
	return tea.Batch(cmds...)
}

func doesFileExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (m model) View() string {
	var b strings.Builder
	normalStyle := lipgloss.NewStyle().Width(60)
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	boldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("420"))
	blazeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("100"))
	b.WriteRune('\n')
	if m.focusScreen == 0 {
		b.WriteString(normalStyle.Render("Looks like this is your first time using this program. I will attempt to create a CSV file in your home directory.\n"))
		b.WriteRune('\n')
		b.WriteString(blueStyle.Render(fmt.Sprintf(`> "%s"`, getFilePath())))
		b.WriteString(normalStyle.Render("\n\nThis should be a one-time setup. Press enter to continue, or ctrl+c to quit at any time."))
	}
	if m.focusScreen == 1 {
		b.WriteString(blueStyle.Render("Press enter to go to the next input. Press the up arrow to go to the previous input.\n"))
		for i := range m.inputs {
			b.WriteRune('\n')
			switch i {
			case 0:
				b.WriteString(boldStyle.Render("app-client-id (this is available in your .env.mk local file)"))
			case 1:
				b.WriteString(boldStyle.Render("Sub Account Email"))
			case 2:
				b.WriteString(boldStyle.Render("Sub Account Password"))
			}

			b.WriteRune('\n')
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
	}
	if m.focusScreen == 2 {
		b.WriteString("Does the following look right?\n")
		b.WriteString(fmt.Sprintf(`
App Client ID: %s
Sub Email: %s
Sub Password: %s

Press %v if yes, or %v to go back.`,
			boldStyle.Render(m.inputs[0].Value()),
			blueStyle.Render(m.inputs[1].Value()),
			blazeStyle.Render(m.inputs[2].Value()),
			boldStyle.Render("y"),
			blueStyle.Render("n")))
	}
	if m.focusScreen == 3 {
		b.WriteString(fmt.Sprintf("A csv file has been created at %s", blueStyle.Render(getFilePath())))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(normalStyle.Render("From now on, you can simply call this program to get the auth token in your terminal."))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(normalStyle.Render(fmt.Sprintf("You can manually edit the CSV file if you want to update the credentials. To go through this whole flow again, you can run the program with the %v flag.", blueStyle.Render("--setup"))))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(normalStyle.Render(fmt.Sprintf("For more information, pass the %v flag to the program.", blueStyle.Render("--help"))))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(normalStyle.Render("You can now exit the program with ctrl+c."))
	}
	return b.String()
}

func createFile(pathName, appClientId, email, password string) {
	f, err := os.OpenFile(pathName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		print(err)
		os.Exit(1)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	headers := []string{"clientId", "email", "password"}
	body := []string{appClientId, email, password}

	writer.Write(headers)
	writer.Write(body)
}

func readFile(pathName string) (appClientId, email, password string) {
	f, err := os.Open(pathName)
	if err != nil {
		log.Fatal("Unable to open file"+pathName, err)
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+pathName, err)
	}
	return records[1][0], records[1][1], records[1][2]
}

func getFilePath() string {
	homeDir, _ := os.UserHomeDir()
	credPath := homeDir + path
	return credPath

}

func getToken() string {
	appClientId, email, password := readFile(getFilePath())
	loginObj, err := aws.GetDetails(appClientId, email, password)
	if err != nil {
		log.Fatal("Couldn't log in with the provided credentials", err)
	}
	return *loginObj.IdToken
}

func main() {
	helpFlag := flag.Bool("help", false, "List help")
	setupFlag := flag.Bool("setup", false, "Trigger Setup")
	copyFlag := flag.Bool("copy", false, "Copies token to clipboard")
	flag.Parse()
	if *helpFlag {
		fmt.Println(`
This is a program to help ease the annoyance of needing to repeatedly
go into the browser devtools to copy the auth token to reuse in Postman.

The simple idea is - on first launch, you are instructed to enter 
several values that will then be saved in a CSV file in your user folder.

On subsequent runs, these values will be passed to AWS' cognito golang 
package to get the IDToken, which can then be used to auth you in Postman.

The CSV file is saved under your home directory, under "~/.swing-jwt-helper.csv"
	
Flags

--help: This screen.

--setup: Initiate the setup screen.

--copy: Copies the token value to your clipboard in addition to printing 
	it in the terminal menu.`)
		os.Exit(1)
	}
	if !doesFileExist(getFilePath()) || *setupFlag {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		token := getToken()
		if *copyFlag {
			clipboard.WriteAll(token)
		}
		fmt.Println(token)
	}
}
