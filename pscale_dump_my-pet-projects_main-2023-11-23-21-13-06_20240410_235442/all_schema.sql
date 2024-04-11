CREATE TABLE `Attraction` (
  `id` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `nameLocal` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `address` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `latitude` double NOT NULL DEFAULT '0',
  `longitude` double NOT NULL DEFAULT '0',
  `isMustSee` tinyint(1) NOT NULL DEFAULT '0',
  `isPredefined` tinyint(1) NOT NULL DEFAULT '0',
  `originalUri` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cityId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `oldId` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `Attraction_cityId_idx` (`cityId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `City` (
  `oldIdForDelete` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `id` bigint NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `alternateNames` text COLLATE utf8mb4_unicode_ci,
  `countryCode` varchar(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `admin1Code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `admin2Code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `admin3Code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `admin4Code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `modificationDate` date NOT NULL,
  `population` int DEFAULT NULL,
  `latitude` double NOT NULL,
  `longitude` double NOT NULL,
  PRIMARY KEY (`id`),
  KEY `City_countryCode_idx` (`countryCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `Color` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `Country` (
  `cca2` varchar(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cca3` varchar(3) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ccn3` varchar(3) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `nameCommon` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `nameOfficial` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `region` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `subregion` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `flagPng` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`cca2`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `Direction` (
  `startPlaceId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `endPlaceId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `directionsData` json NOT NULL,
  UNIQUE KEY `Direction_startPlaceId_endPlaceId_key` (`startPlaceId`,`endPlaceId`),
  KEY `Direction_startPlaceId_idx` (`startPlaceId`),
  KEY `Direction_endPlaceId_idx` (`endPlaceId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `Itinerary` (
  `id` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tripId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `order` int NOT NULL,
  `colorId` int NOT NULL,
  PRIMARY KEY (`id`),
  KEY `Itinerary_tripId_idx` (`tripId`),
  KEY `Itinerary_colorId_idx` (`colorId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `ItineraryPlace` (
  `id` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `itineraryId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `attractionId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `order` int NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ItineraryPlace_itineraryId_attractionId_key` (`itineraryId`,`attractionId`),
  KEY `ItineraryPlace_itineraryId_idx` (`itineraryId`),
  KEY `ItineraryPlace_attractionId_idx` (`attractionId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `Trip` (
  `id` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `startDate` datetime(6) DEFAULT NULL,
  `endDate` datetime(6) DEFAULT NULL,
  `oldId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `TripDestinations` (
  `tripId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `countryId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`tripId`,`countryId`),
  UNIQUE KEY `TripDestinations_tripId_countryId_key` (`tripId`,`countryId`),
  KEY `TripDestinations_tripId_idx` (`tripId`),
  KEY `TripDestinations_countryId_idx` (`countryId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `breweries` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `website` varchar(100) DEFAULT NULL,
  `geo_id` bigint NOT NULL,
  `old_id` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_name_geo_id` (`name`,`geo_id`),
  KEY `fk_geo_id` (`geo_id`),
  CONSTRAINT `fk_geo_id` FOREIGN KEY (`geo_id`) REFERENCES `City` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=683 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `beers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `brand` varchar(100) NOT NULL,
  `type` varchar(100) DEFAULT NULL,
  `style` varchar(100) NOT NULL,
  `brewery_id` bigint DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '0',
  `created_at` datetime NOT NULL,
  `updated_at` datetime DEFAULT NULL,
  `old_image_ids` varchar(200) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_brewery_id` (`brewery_id`),
  CONSTRAINT `fk_brewery_id` FOREIGN KEY (`brewery_id`) REFERENCES `breweries` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2168 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `trip_medias` (
  `id` varchar(191) NOT NULL,
  `trip_id` varchar(191) NOT NULL,
  `name` varchar(191) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
