# msg-assembler
I attempted the Multi-threading extra credit.

## Design
The data model I chose for handling the fragments of a message is multiple
hash maps and a binary tree. When a fragment is received by the server.go module
it uses the `CreateFragment` function in fragment.go to create a fragment. After
creating the fragment, it is handed off to the `MsgHandler` to add it to the
data model. `MsgHandler` wraps the data model with a `sync.Mutex` to make sure
only a single go routine can access the data model at one time. `MsgHandler` also
implements the clean up functionality.

### Clean Up
To implement the 30 second timeout waiting for the entire message I use a
`time.AfterFunc` to execute a routine to remove the message and fragments
from the data model. A clean up timer is created when the first fragment of a
message is received. One clean up timer exists for each unique transaction ID
of a message.

### Data Model
The msg.go file implements most of the in memory data model. I use a hash map and
a binary tree to solve two problems. The hash map solves quickly maping a fragment
with its message. This hash map is implemented in the `MsgHandler` to find the right
`Msg` when a fragment is received. `Msg` also uses a map to determine if the fragment
is a duplicate. The binary tree used by `Msg` is to solve the issue of finding holes
in a message. My solution to this was to sort all the fragments by their offset and
then look at each fragment to determine if a fragment is missing. Instead of keeping
all the fragments out of order an then using golang's sort routine when looking for holes
which I assume is O(n log n). Using an array would probably require many reallocations
and copying of the fragment pointers. This probably isn't a huge deal but if
the messages contained millions of fragments it could be an issue. The binary
tree will keep the fragments in sorted order as they arrive. This would not be a good
design decision if the fragments arrived in order because the tree would be very right
side heavy and insertions would take O(n). A self balancing tree could be used to handle
that though.

### Testing
I tried to use TDD and write unit tests as I went. The server.go code is lacking in its
testing coverage. This is mainly because of the difficulty testing the go routines for
handling reading the data off the UDP port. I tried to mock net so I could control what
data was sent without needing an actual UDP client but I was getting a lot of deadlocks.

## Assumptions
For the hole identification functionality, if the final fragment hasn't been
received the server will print a hole at the offset where the greatest offset exists.
The greatest offset is the largest received fragment (LRF) offset +
LRF's data length. If a fragment with offset 0 hasn't been received the server will
also print a hole for that location as well.

### Bad Data
The server does not handle malformed packets very well. If for example the client sent
fragments that overlapped the server would not handle this very well. The way I keep
track of whether all the fragments have been received is by keeping a running total of
all the data and comparing that with the last fragment's offset + data length. If the
last fragment is never received then I also take that into consideration with a flag.
To solve this one could use another data structure to determine if a fragment received
overlaps with an already existing fragment.

