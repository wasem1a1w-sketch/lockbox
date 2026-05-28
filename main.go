package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	addCmd := flag.String("add", "", "Register a credential: 'domain.com:username:password'")
	listCmd := flag.Bool("list", false, "Decrypt and list all records inside the vault")
	editCmd := flag.String("edit", "", "Update password for a credential: 'index:newpassword'")
	reorderCmd := flag.String("reorder", "", "Reorder a credential: 'old_index:new_index'")
	deleteCmd := flag.Int("delete", -1, "Delete a credential by its index")
	genCmd := flag.Int("gen", 0, "Generate a strong password of the specified length")
	flag.Parse()

	storage := NewSecureStorage("vault.lock")

	key, salt, vault, err := authenticate(storage)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch {
	case *addCmd != "":
		handleAdd(&vault, storage, key, salt, *addCmd)

	case *listCmd:
		handleList(vault)

	case *editCmd != "":
		handleEdit(&vault, storage, key, salt, *editCmd)

	case *reorderCmd != "":
		handleReOrderFlag(&vault, storage, key, salt, *reorderCmd)

	case *deleteCmd >= 0:
		handleDelete(&vault, storage, key, salt, *deleteCmd)

	case *genCmd > 0:
		handleGenerate(*genCmd)

	default:
		fmt.Println("ℹ️ Action required. Pass a flag: --add, --list, --edit, --delete, or --gen")
	}
}

func authenticate(storage *SecureStorage) ([]byte, []byte, Vault, error) {
	salt, err := storage.GetOrGenerateSalt()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("🚨 Initialization error: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("🔑 Enter Vault Master Password: ")
	masterInput, _ := reader.ReadString('\n')
	masterInput = strings.TrimSpace(masterInput)

	if masterInput == "" {
		return nil, nil, nil, fmt.Errorf("🚨 Error: Master password cannot be blank.")
	}

	key := DeriveKey(masterInput, salt)

	vault, err := storage.Load(key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("❌ Authentication Failed: %v", err)
	}

	fmt.Println("🔓 Vault unlocked successfully.")
	return key, salt, vault, nil
}

func handleAdd(vault *Vault, storage *SecureStorage, key, salt []byte, raw string) {
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		fmt.Println("🚨 Formatting error! Please match: --add 'service.com:user:pass'")
		return
	}

	vault.Add(parts[0], parts[1], parts[2])
	if err := storage.Save(*vault, key, salt); err != nil {
		fmt.Printf("🚨 Failed to commit changes to file: %v\n", err)
		return
	}
	fmt.Printf("🛡️ Account profile successfully encrypted for target: %s\n", parts[0])
}

func handleList(vault Vault) {
	if len(vault) == 0 {
		fmt.Println("📭 Your cryptographic vault is empty.")
		return
	}

	sort.Slice(vault, func(i, j int) bool {
		return vault[i].Index < vault[j].Index
	})

	fmt.Println("\n🔐 --- SECURE DECRYPTED ACCOUNTS ---")
	for _, cred := range vault {
		fmt.Printf("[%d] Context: %s | Account ID: %s | Master Key: %s (Saved: %s)\n",
			cred.Index, cred.Account, cred.Username, cred.Password, cred.SavedAt.Format("2006-01-02 15:04"))
	}
}

func handleEdit(vault *Vault, storage *SecureStorage, key, salt []byte, raw string) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		fmt.Println("🚨 Formatting error! Please match: --edit 'index:newpassword'")
		return
	}

	index, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Println("🚨 Error: index must be a number")
		return
	}

	if err := vault.EditPassword(index, parts[1]); err != nil {
		fmt.Printf("🚨 %v\n", err)
		return
	}

	if err := storage.Save(*vault, key, salt); err != nil {
		fmt.Printf("🚨 Failed to commit changes to file: %v\n", err)
		return
	}
	fmt.Printf("🛡️ Password updated for credential index %d\n", index)
}

func handleDelete(vault *Vault, storage *SecureStorage, key, salt []byte, index int) {
	if err := vault.Delete(index); err != nil {
		fmt.Printf("🚨 %v\n", err)
		return
	}

	if err := storage.Save(*vault, key, salt); err != nil {
		fmt.Printf("🚨 Failed to commit changes to file: %v\n", err)
		return
	}
	fmt.Printf("🗑️ Credential with index %d deleted successfully\n", index)
}

func handleReOrder(vault *Vault, storage *SecureStorage, key, salt []byte, old_index int, new_index int) {
	if err := vault.ReOrder(old_index, new_index); err != nil {
		fmt.Printf("🚨 %v\n", err)
		return
	}

	if err := storage.Save(*vault, key, salt); err != nil {
		fmt.Printf("🚨 Failed to commit changes to file: %v\n", err)
		return
	}
	fmt.Printf("🛡️ Credential reordered for credential from index %d to index %d\n", old_index, new_index)
}

func handleReOrderFlag(vault *Vault, storage *SecureStorage, key, salt []byte, raw string) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		fmt.Println("🚨 Formatting error! Please match: --reorder 'old_index:new_index'")
		return
	}

	oldIndex, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Println("🚨 Error: old_index must be a number")
		return
	}

	newIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("🚨 Error: new_index must be a number")
		return
	}

	handleReOrder(vault, storage, key, salt, oldIndex, newIndex)
}

func handleGenerate(length int) {
	password, err := GeneratePassword(length)
	if err != nil {
		fmt.Printf("🚨 Failed to generate password: %v\n", err)
		return
	}
	fmt.Printf("🔐 Generated Password: %s\n", password)
}
