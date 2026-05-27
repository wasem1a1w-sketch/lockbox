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
