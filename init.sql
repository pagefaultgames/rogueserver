CREATE TABLE accounts (
  uuid BINARY(16) NOT NULL PRIMARY KEY,
  username VARCHAR(16) NOT NULL,
  hash BINARY(32) NOT NULL,
  salt BINARY(16) NOT NULL, 
  registered DATETIME NOT NULL DEFAULT NOW(),
  lastActivity DATETIME NOT NULL DEFAULT NOW(),
  lastLoggedIn DATETIME NULL,
  INDEX (lastActivity),
  INDEX (username),
  UNIQUE (username)
);

CREATE TABLE sessions (
  token BINARY(32) NOT NULL PRIMARY KEY,
  uuid BINARY(16) NOT NULL,
  expire DATETIME NOT NULL,
  FOREIGN KEY (uuid) REFERENCES accounts(uuid) ON DELETE CASCADE
);

CREATE TABLE dailyRuns (
  seed VARCHAR(255) NOT NULL,
  date DATE NOT NULL PRIMARY KEY
);

CREATE TABLE accountDailyRuns (
  uuid BINARY(16) NOT NULL PRIMARY KEY,
  date DATE NOT NULL,
  score INT NOT NULL,
  wave INT NOT NULL,
  timestamp DATETIME,
  FOREIGN KEY (date) REFERENCES dailyRuns(date),
  FOREIGN KEY (uuid) REFERENCES accounts(uuid) ON DELETE CASCADE,
  UNIQUE (uuid, date)
);

CREATE TABLE seedCompletions (
  uuid BINARY(16) NOT NULL PRIMARY KEY,
  seed VARCHAR(255) NOT NULL,
  mode INT NOT NULL,
  timestamp DATETIME NOT NULL DEFAULT NOW(),
  FOREIGN KEY (uuid) REFERENCES accounts(uuid) ON DELETE CASCADE,
  UNIQUE (seed, mode, uuid)
);
