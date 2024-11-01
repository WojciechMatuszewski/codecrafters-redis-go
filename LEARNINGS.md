# Learnings

- The `conn.Read` can throw the `IOF` error when the client closes the connection or when there is no more data to be read from the connection.

  - This cna happen when you _read_ from the connection in a loop, but the client closed the connection

- For reading/writing data, you might consider using the `bufio.Reader` and `bufio.Writer` respectively.

  - The _reader_ and the _writer_ hold internal state. Depending on how much you want to read, you **might need to configure the "sliding window" the _reader_ uses**.

- There are multiple "flavours" of _mutexes_ in go.

  - There is the "basic" `Mutex` which will **lock both readers and writers** when lock is acquired.

  - The is the `RWMutex` which **will lock both readers and writers when writing, but allows access by multiple readers**.

  - Depending on your use-case, for example in situations where there are a lot of reads, it might be worth using `RWMutex` over the `Mutex`.

- Use the `%q` _format verb_ to display all whitespace encoded as characters.

  - Very useful for looking how the Redis input command is structured!

- **Shutting down the server gracefully is, to me, surprisingly challenging**.

  - A lot of calls are blocking, but the fact that they are, is never documented anywhere?

    - The flip side is that you write the async code as if it was synchronous.
