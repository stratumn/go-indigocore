// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package fossilizertesting

import (
	"github.com/stratumn/go/fossilizer"
)

// MockAdapter is used to mock a fossilizer.
//
// It implements github.com/stratumn/go/fossilizer.Adapter.
type MockAdapter struct {
	// The mock for the GetInfo function.
	MockGetInfo MockGetInfo

	// The mock for the AddResultChan function.
	MockAddResultChan MockAddResultChan

	// The mock for the Fossilize function.
	MockFossilize MockFossilize
}

// MockGetInfo mocks the GetInfo function.
type MockGetInfo struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() (interface{}, error)
}

// MockAddResultChan mocks the AddResultChan function.
type MockAddResultChan struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []chan *fossilizer.Result

	// The last channel that was passed.
	LastCalledWith chan *fossilizer.Result

	// An optional implementation of the function.
	Fn func(chan *fossilizer.Result)
}

// MockFossilize mocks the Fossilize function.
type MockFossilize struct {
	// The number of times the function was called.
	CalledCount int

	// The data that was passed to each call.
	CalledWithData [][]byte

	// The meta that was passed to each call.
	CalledWithMeta [][]byte

	// The last data that was passed.
	LastCalledWithData []byte

	// The last meta that was passed.
	LastCalledWithMeta []byte

	// An optional implementation of the function.
	Fn func([]byte, []byte) error
}

// GetInfo implements github.com/stratumn/go/fossilizer.Adapter.GetInfo.
func (a *MockAdapter) GetInfo() (interface{}, error) {
	a.MockGetInfo.CalledCount++

	if a.MockGetInfo.Fn != nil {
		return a.MockGetInfo.Fn()
	}

	return nil, nil
}

// AddResultChan implements github.com/stratumn/go/fossilizer.Adapter.AddResultChan.
func (a *MockAdapter) AddResultChan(resultChan chan *fossilizer.Result) {
	a.MockAddResultChan.CalledCount++
	a.MockAddResultChan.CalledWith = append(a.MockAddResultChan.CalledWith, resultChan)
	a.MockAddResultChan.LastCalledWith = resultChan

	if a.MockAddResultChan.Fn != nil {
		a.MockAddResultChan.Fn(resultChan)
	}
}

// Fossilize implements github.com/stratumn/go/fossilizer.Adapter.Fossilize.
func (a *MockAdapter) Fossilize(data []byte, meta []byte) error {
	a.MockFossilize.CalledCount++
	a.MockFossilize.CalledWithData = append(a.MockFossilize.CalledWithData, data)
	a.MockFossilize.LastCalledWithData = data
	a.MockFossilize.CalledWithMeta = append(a.MockFossilize.CalledWithMeta, meta)
	a.MockFossilize.LastCalledWithMeta = meta

	if a.MockFossilize.Fn != nil {
		return a.MockFossilize.Fn(data, meta)
	}

	return nil
}
