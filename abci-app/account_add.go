/*
 * Copyright (C) 2018 eeonevision
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package app

import (
	"errors"

	"github.com/eeonevision/anychaindb/crypto"
	"github.com/eeonevision/anychaindb/state"
	"github.com/eeonevision/anychaindb/transaction"
)

func checkAccountAddTransaction(tx *transaction.Transaction, s *state.State) error {
	data := &state.Account{}
	_, err := data.UnmarshalMsg(tx.Data)
	if err != nil {
		return err
	}
	if s.HasAccount(data.ID) {
		return errors.New("account exists")
	}
	if _, err := crypto.NewFromStrings(data.PubKey, ""); err != nil {
		return err
	}
	return nil
}

func deliverAccountAddTransaction(tx *transaction.Transaction, s *state.State) error {
	data := &state.Account{}
	_, err := data.UnmarshalMsg(tx.Data)
	if err != nil {
		return err
	}
	return s.AddAccount(data)
}
