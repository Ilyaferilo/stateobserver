# state observer

StateObserver used to synchronize the local(cached) state of the remote object with the real state. 
Synchronization is provided by constantly polling the object using the `getState` interface function.
The functions for getting and setting object states are blocking and are always called from a single goroutine that is created by the `Start` method
