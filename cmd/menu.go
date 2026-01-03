package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Make sure CompareRepos is imported from compare.go
// import "github.com/agnivo988/Repo-lyzer/cmd"

// RunMenu launches the interactive menu
func RunMenu() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nWelcome to Repo-lyzer!")
		fmt.Println("1. Analyze a repository")
		fmt.Println("2. Compare two repositories")
		fmt.Println("3. Exit")
		fmt.Print("Enter choice (1-3): ")

		choiceRaw, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(choiceRaw)

		switch choice {
		case "1":
			fmt.Println("Analyze feature coming soon…") // Replace with your analyze logic

		case "2":
			// Ask for repo inputs
			fmt.Print("Enter first repository (owner/repo): ")
			r1, _ := reader.ReadString('\n')
			r1 = strings.TrimSpace(r1)

			fmt.Print("Enter second repository (owner/repo): ")
			r2, _ := reader.ReadString('\n')
			r2 = strings.TrimSpace(r2)

			// Call compare logic
			err := CompareRepos(r1, r2)
			if err != nil {
				fmt.Println("Error:", err)
			}

		case "3":
			fmt.Println("Goodbye!")
			os.Exit(0)

		default:
			fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
		}
	}
}
