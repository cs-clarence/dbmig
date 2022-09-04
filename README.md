# dbmig

It creates migration files.
It doesn't create the content of those migration files.
It merely creates the files and you fill the content of those files.

## Usage:

- Initialize the current working directory

  ```bash
  go run -mod=mod github.com/cs-clarence/dbmig init .
  ```

  or (if you `go install` the package)

  ```bash
  dbmig init .
  ```

- Creating a new migration (substitute `<name>`)

  ```bash
  go run -mod=mod github.com/cs-clarence/dbmig new <name>
  ```

  or (if you `go install` the package)

  ```bash
  dbmig new <name>
  ```
