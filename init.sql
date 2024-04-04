CREATE TABLE `accounts` (
  `uuid` binary(16) NOT NULL,
  `username` varchar(16) NOT NULL,
  `hash` binary(32) NOT NULL,
  `salt` binary(16) NOT NULL,
  `registered` timestamp NOT NULL,
  `lastLoggedIn` timestamp NULL DEFAULT NULL,
  `lastActivity` timestamp NULL DEFAULT NULL,
  `banned` tinyint(2) NOT NULL DEFAULT 0,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci


CREATE TABLE `sessions` (
  `token` binary(32) NOT NULL,
  `uuid` binary(16) NOT NULL,
  `expire` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`token`),
  KEY `uuid` (`uuid`),
  CONSTRAINT `uuid` FOREIGN KEY (`uuid`) REFERENCES `accounts` (`uuid`) ON DELETE NO ACTION ON UPDATE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci


CREATE TABLE `seedCompletions` (
  `uuid` binary(16) NOT NULL,
  `seed` char(24) NOT NULL,
  `mode` int(11) NOT NULL DEFAULT 0,
  `timestamp` timestamp NOT NULL,
  `score` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`uuid`,`seed`),
  KEY `uuid` (`uuid`),
  CONSTRAINT `seedCompletions_ibfk_1` FOREIGN KEY (`uuid`) REFERENCES `accounts` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci


CREATE TABLE `dailyRuns` (
  `date` date NOT NULL,
  `seed` char(24) NOT NULL,
  PRIMARY KEY (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci


CREATE TABLE `accountDailyRuns` (
  `uuid` binary(16) NOT NULL,
  `date` date NOT NULL,
  `timestamp` timestamp NOT NULL,
  `score` int(11) NOT NULL DEFAULT 0,
  `wave` int(11) NOT NULL,
  PRIMARY KEY (`uuid`,`date`),
  KEY `uuid` (`uuid`),
  KEY `date` (`date`),
  CONSTRAINT `accountDailyRuns_ibfk_1` FOREIGN KEY (`uuid`) REFERENCES `accounts` (`uuid`),
  CONSTRAINT `accountDailyRuns_ibfk_2` FOREIGN KEY (`date`) REFERENCES `dailyRuns` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci