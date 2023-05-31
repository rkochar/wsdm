# MySQL Info for Lockmaster
The database can be created using:
```
CREATE DATABASE saga_log CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```


## `sagas` table
The `sagas` table holds the Saga IDs and the timestamps when they were created.
Rows look like:
ID (int)
timestamp (timestamp)

#### Create Table Query
```
CREATE TABLE IF NOT EXISTS sagas (
	ID INT AUTO_INCREMENT PRIMARY KEY,
	timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;
```

## `messsages` table
The `messsages` table contains the specific SAGA messages.
Rows look like:
ID (int)
saga_id (int)
message_type (int) FK to message_types.ID
message_event (int) FK to message_events.ID
saga_contents (text)
timestamp (timestamp)

#### Create Table Query
```
CREATE TABLE IF NOT EXISTS messages (
        ID INT AUTO_INCREMENT PRIMARY KEY,
        saga_id INT,
        message_type INT,
        message_event INT,
        saga_contents TEXT,
        timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (saga_id) REFERENCES sagas(ID),
        FOREIGN KEY (message_type) REFERENCES message_types(ID),
        FOREIGN KEY (message_event) REFERENCES message_events(ID)
) ENGINE=InnoDB;
```

## `message_types` table
Rows look like:
ID (int) PK
type (varchar(10))

#### Create Table Query
```
CREATE TABLE IF NOT EXISTS message_types (
        ID INT PRIMARY KEY,
        type varchar(10)
) ENGINE=InnoDB;

INSERT INTO message_types (type)
VALUES 
  ("START"),
  ("END"),
  ("ABORT")
WHERE NOT EXISTS (SELECT * FROM message_types);
```

## `message_events` table
Rows look like:
ID (int) PK
event (varchar(32))

#### Create Table Query
```
CREATE TABLE IF NOT EXISTS message_events (
        ID INT PRIMARY KEY,
        event varchar(32)
) ENGINE=InnoDB;

INSERT INTO message_events (event)
VALUES 
  ("MAKE-PAYMENT"),
  ("CANCEL-PAYMENT"),
  ("CHECKOUT-SAGA"),
  ("CANCEL-SAGA"),
  ("SUBTRACT-STOCK"),
  ("READD-STOCK"),
  ("UPDATE-ORDER")
WHERE NOT EXISTS (SELECT * FROM message_events);
```
