CREATE TABLE `TripDestinations` (
  `tripId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `countryId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`tripId`,`countryId`),
  UNIQUE KEY `TripDestinations_tripId_countryId_key` (`tripId`,`countryId`),
  KEY `TripDestinations_tripId_idx` (`tripId`),
  KEY `TripDestinations_countryId_idx` (`countryId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
