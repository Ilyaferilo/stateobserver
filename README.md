# state observer

[![Go](https://github.com/Ilyaferilo/state_observer/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/Ilyaferilo/state_observer/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/Ilyaferilo/state_observer/branch/master/graph/badge.svg)](https://codecov.io/gh/Ilyaferilo/state_observer)


StateObserver used to synchronize the local(cached) state of the remote object with the real state. 
Synchronization is provided by constantly polling the object using the `getState` interface function.
The functions for getting and setting object states are blocking and are always called from a single goroutine that is created by the `Start` method
