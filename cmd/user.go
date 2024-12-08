package cmd

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"codeinstyle.io/captain/config"
	"codeinstyle.io/captain/db"
	"codeinstyle.io/captain/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func getValidInput(prompt string, validator func(string) error) string {
	for {
		var input string
		fmt.Print(prompt)
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Printf("Error: failed to read input: %v. Please try again.\n", err)
			continue
		}

		if err := validator(input); err != nil {
			fmt.Printf("Error: %v. Please try again.\n", err)
			continue
		}
		return input
	}
}

func getValidPassword(prompt string) string {
	fmt.Println("\nPassword requirements:")
	fmt.Println("- At least 8 characters long")
	fmt.Println("- At least one uppercase letter")
	fmt.Println("- At least one lowercase letter")
	fmt.Println("- At least one number")
	fmt.Println("- At least one special character (!@#$%^&*(),.?\":{}|<>)")
	fmt.Println()

	for {
		var password string
		passwordBytes, err := readPassword(prompt)
		if err != nil {
			panic(err)
		}
		password = string(passwordBytes)
		fmt.Println() // Add newline after password input

		if err := ValidatePassword(password); err != nil {
			fmt.Printf("Error: %v. Please try again.\n", err)
			continue
		}
		return password
	}
}

func readPassword(prompt string) ([]byte, error) {
	fmt.Fprint(os.Stderr, prompt)
	return term.ReadPassword(int(syscall.Stdin))
}

func CreateUser(cmd *cobra.Command, args []string) {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	database, err := db.InitDB(cfg)

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	firstName := getValidInput("First Name: ", ValidateFirstName)
	lastName := getValidInput("Last Name: ", ValidateLastName)
	email := getValidInput("Email: ", ValidateEmail)
	password := getValidPassword("Password: ")

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("Failed to hash password: %v\n", err)
		return
	}

	user := &db.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  hashedPassword,
	}

	if err := db.CreateUser(database, user); err != nil {
		log.Printf("Failed to create user: %v\n", err)
		return
	}

	fmt.Println("User created successfully")
}

func UpdateUserPassword(cmd *cobra.Command, args []string) {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	email := getValidInput("Email: ", ValidateEmail)

	var user db.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		fmt.Println("User not found")
		return
	}

	for {
		fmt.Print("Old Password: ")
		oldPasswordBytes, _ := term.ReadPassword(0)
		fmt.Println()

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), oldPasswordBytes); err != nil {
			fmt.Println("Incorrect password. Please try again.")
			continue
		}
		break
	}

	newPassword := getValidPassword("New Password: ")

	for {
		fmt.Print("Confirm Password: ")
		confirmBytes, _ := term.ReadPassword(0)
		fmt.Println()
		confirmPassword := string(confirmBytes)

		if newPassword != confirmPassword {
			fmt.Println("Passwords don't match. Please try again.")
			newPassword = getValidPassword("New Password: ")
			continue
		}
		break
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		log.Printf("Failed to hash password: %v\n", err)
		return
	}
	user.Password = hashedPassword

	if err := database.Save(&user).Error; err != nil {
		log.Printf("Failed to update password: %v\n", err)
		return
	}

	fmt.Println("Password updated successfully")
}
