
# migrator

[![Build Status](https://travis-ci.org/altipla-consulting/migrator.svg?branch=master)](https://travis-ci.org/altipla-consulting/migrator)

Apply SQL migrations from a folder.


### Install

```shell
go get github.com/altipla-consulting/migrator/cmd/init-migrator
go get github.com/altipla-consulting/migrator/cmd/migrator
```


### Usage

Run the initialize one time per database server to create the state needed to track the applied migrations:

```shell
init-migrator -user foo -password bar -address mysql.example.com
```

Then run the application as many times as yu want in your CI process or machine. Each time the app is called only the new migrations will be applied.

```shell
migrator -user foo -password bar -address mysql.example.com -directory ./migrations
```

The folder `./migrations` of the example commands should contain one file per migration. Each file can have multiple SQL statements and every one of them should finish with a `;` character.

The first statement should always be `USE mydbname;` selecting the database you want to use for the rest of the file. If you want to migrate two different databases use two different files, `USE` can only be used once per file.

Example migrations:

```sql
USE information_schema;

CREATE SCHEMA foo;
```

```sql
USE foo;

CREATE TABLE images (
  id INT(11) NOT NULL AUTO_INCREMENT,
  filename VARCHAR(255) NOT NULL,

  revision INT(11) NOT NULL,
  
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
```


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using `gofmt`.


### Running tests

Run a full test suite from a blank state:

```shell
make test
```


### License

[MIT License](LICENSE)
