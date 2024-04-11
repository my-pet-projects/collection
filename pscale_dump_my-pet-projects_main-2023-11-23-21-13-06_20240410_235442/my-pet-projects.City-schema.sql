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
