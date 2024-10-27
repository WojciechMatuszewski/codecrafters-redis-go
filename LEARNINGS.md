# Learnings

- The `conn.Read` can throw the `IOF` error when the client closes the connection or when there is no more data to be read from the connection.

  - This cna happen when you _read_ from the connection in a loop, but the client closed the connection
