package main

import (
	"fmt"
	"time"
)

type Credential struct {
	Index    int       `json:"index"`
	Account  string    `json:"account"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	SavedAt  time.Time `json:"saved_at"`
}

type Vault []Credential

func (v *Vault) Add(account, username, password string) {
	next := 1
	for _, c := range *v {
		if c.Index >= next {
			next = c.Index + 1
		}
	}
	*v = append(*v, Credential{
		Index:    next,
		Account:  account,
		Username: username,
		Password: password,
		SavedAt:  time.Now(),
	})
}

func (v *Vault) Delete(index int) error {
	for i, c := range *v {
		if c.Index == index {
			*v = append((*v)[:i], (*v)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("credential with index %d not found", index)
}

func (vault *Vault) ReOrder(old_index int, new_index int) error {
	if old_index == new_index {
		return nil
	}
	if old_index < 1 || new_index < 1 {
		return fmt.Errorf("indices must be above zero")
	}

	pos := -1
	for i, value := range *vault {
		if value.Index == old_index {
			pos = i
			break
		}
	}
	if pos == -1 {
		return fmt.Errorf("credential with index %d not found", old_index)
	}

	(*vault)[pos].Index = new_index

	if new_index < old_index {
		for i, value := range *vault {
			if i != pos && value.Index >= new_index && value.Index < old_index {
				(*vault)[i].Index++
			}
		}
	} else {
		for i, value := range *vault {
			if i != pos && value.Index > old_index && value.Index <= new_index {
				(*vault)[i].Index--
			}
		}
	}

	return nil
}

func (v *Vault) EditPassword(index int, password string) error {
	for i, c := range *v {
		if c.Index == index {
			(*v)[i].Password = password
			(*v)[i].SavedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("credential with index %d not found", index)
}
