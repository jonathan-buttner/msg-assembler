# msg-assembler

## Assumptions
For the hole identification functionality, if the final fragment hasn't been
received the server will print a hole at the offset where the greatest offset.
The greatest offset is the largest received fragment (LRF) offset +
LRF's data length.

