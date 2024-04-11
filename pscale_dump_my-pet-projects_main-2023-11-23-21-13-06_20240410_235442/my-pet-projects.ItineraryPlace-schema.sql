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
