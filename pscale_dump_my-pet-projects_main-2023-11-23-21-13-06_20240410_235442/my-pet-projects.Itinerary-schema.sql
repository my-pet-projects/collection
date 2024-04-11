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
