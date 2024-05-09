/*RUN THESE IN ORDER TO BUILD YOUR DATABASE*/
DROP TABLE IF EXISTS `dailyRuns`;
CREATE TABLE `dailyRuns` (
  `date` date NOT NULL,
  `seed` char(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  PRIMARY KEY (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `accounts`;
CREATE TABLE `accounts` (
  `uuid` binary(16) NOT NULL,
  `username` varchar(16) NOT NULL,
  `hash` binary(32) NOT NULL,
  `salt` binary(16) NOT NULL,
  `registered` timestamp NOT NULL,
  `lastLoggedIn` timestamp NULL DEFAULT NULL,
  `lastActivity` timestamp NULL DEFAULT NULL,
  `banned` tinyint(1) NOT NULL DEFAULT 0,
  `trainerId` smallint(5) unsigned DEFAULT 0,
  `secretId` smallint(5) unsigned DEFAULT 0,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `accountCompensations`;
CREATE TABLE `accountCompensations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uuid` binary(16) NOT NULL,
  `voucherType` int(11) NOT NULL,
  `count` int(11) NOT NULL,
  `claimed` bit(1) NOT NULL DEFAULT b'0',
  PRIMARY KEY (`id`),
  KEY `uuid` (`uuid`),
  CONSTRAINT `accountCompensations_ibfk_1` FOREIGN KEY (`uuid`) REFERENCES `accounts` (`uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=395447 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


DROP TABLE IF EXISTS `accountDailyRuns`;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `accountStats`;
CREATE TABLE `accountStats` (
  `uuid` binary(16) NOT NULL,
  `playTime` int(11) NOT NULL DEFAULT 0,
  `battles` int(11) NOT NULL DEFAULT 0,
  `classicSessionsPlayed` int(11) NOT NULL DEFAULT 0,
  `sessionsWon` int(11) NOT NULL DEFAULT 0,
  `highestEndlessWave` int(11) NOT NULL DEFAULT 0,
  `highestLevel` int(11) NOT NULL DEFAULT 0,
  `pokemonSeen` int(11) NOT NULL DEFAULT 0,
  `pokemonDefeated` int(11) NOT NULL DEFAULT 0,
  `pokemonCaught` int(11) NOT NULL DEFAULT 0,
  `pokemonHatched` int(11) NOT NULL DEFAULT 0,
  `eggsPulled` int(11) NOT NULL DEFAULT 0,
  `regularVouchers` int(11) NOT NULL DEFAULT 0,
  `plusVouchers` int(11) NOT NULL DEFAULT 0,
  `premiumVouchers` int(11) NOT NULL DEFAULT 0,
  `goldenVouchers` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`uuid`),
  KEY `uuid` (`uuid`),
  CONSTRAINT `accountStats_ibfk_1` FOREIGN KEY (`uuid`) REFERENCES `accounts` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


DROP TABLE IF EXISTS `dailyRunCompletions`;
CREATE TABLE `dailyRunCompletions` (
  `uuid` binary(16) NOT NULL,
  `seed` char(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `mode` int(11) NOT NULL DEFAULT 0,
  `timestamp` timestamp NOT NULL,
  `score` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`uuid`,`seed`),
  KEY `uuid` (`uuid`),
  CONSTRAINT `dailyRunCompletions_ibfk_1` FOREIGN KEY (`uuid`) REFERENCES `accounts` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;




DROP TABLE IF EXISTS `sessions`;
CREATE TABLE `sessions` (
  `token` binary(32) NOT NULL,
  `uuid` binary(16) NOT NULL,
  `expire` timestamp NULL DEFAULT NULL,
  `active` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`token`),
  KEY `uuid` (`uuid`),
  CONSTRAINT `sessions_ibfk_1` FOREIGN KEY (`uuid`) REFERENCES `accounts` (`uuid`) ON DELETE NO ACTION ON UPDATE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
