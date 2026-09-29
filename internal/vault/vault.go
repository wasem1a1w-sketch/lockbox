package vault

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Credential struct {
	ID        string    `json:"id"`
	Account   string    `json:"account"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	SavedAt   time.Time `json:"saved_at"`
	SortOrder int       `json:"sort_order"`
}

type Vault []Credential

func (v *Vault) Add(account, username, password string) {
	nextOrder := 1
	for _, c := range *v {
		if c.SortOrder >= nextOrder {
			nextOrder = c.SortOrder + 1
		}
	}
	*v = append(*v, Credential{
		ID:        uuid.New().String(),
		Account:   account,
		Username:  username,
		Password:  password,
		SavedAt:   time.Now(),
		SortOrder: nextOrder,
	})
}

func (v *Vault) DeleteByID(id string) error {
	for i, c := range *v {
		if c.ID == id {
			*v = append((*v)[:i], (*v)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("credential with id %s not found", id)
}

func (v *Vault) DeleteByIndex(index int) error {
	if index < 1 || index > len(*v) {
		return fmt.Errorf("invalid index %d", index)
	}
	*v = append((*v)[:index-1], (*v)[index:]...)
	return nil
}

func (v *Vault) EditPasswordByID(id, password string) error {
	for i, c := range *v {
		if c.ID == id {
			(*v)[i].Password = password
			(*v)[i].SavedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("credential with id %s not found", id)
}

func (v *Vault) EditPasswordByIndex(index int, password string) error {
	if index < 1 || index > len(*v) {
		return fmt.Errorf("invalid index %d", index)
	}
	(*v)[index-1].Password = password
	(*v)[index-1].SavedAt = time.Now()
	return nil
}

func (v *Vault) ReOrder(oldIndex, newIndex int) error {
	if oldIndex == newIndex {
		return nil
	}
	if oldIndex < 1 || newIndex < 1 || oldIndex > len(*v) || newIndex > len(*v) {
		return fmt.Errorf("indices must be between 1 and %d", len(*v))
	}

	item := (*v)[oldIndex-1]
	*v = append((*v)[:oldIndex-1], (*v)[oldIndex:]...)

	if newIndex > len(*v) {
		newIndex = len(*v) + 1
	}
	*v = append((*v)[:newIndex-1], append([]Credential{item}, (*v)[newIndex-1:]...)...)

	for i := range *v {
		(*v)[i].SortOrder = i + 1
	}

	return nil
}

func (v Vault) FindByID(id string) *Credential {
	for i := range v {
		if v[i].ID == id {
			return &v[i]
		}
	}
	return nil
}

func (v Vault) FindByIndex(index int) *Credential {
	if index < 1 || index > len(v) {
		return nil
	}
	return &v[index-1]
}

func (v Vault) Filter(search, username string) Vault {
	if search == "" && username == "" {
		return v
	}
	var result Vault
	for _, c := range v {
		match := true
		if search != "" {
			match = match && (contains(c.Account, search) || contains(c.Username, search) || contains(c.Password, search))
		}
		if username != "" {
			match = match && contains(c.Username, username)
		}
		if match {
			result = append(result, c)
		}
	}
	return result
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}