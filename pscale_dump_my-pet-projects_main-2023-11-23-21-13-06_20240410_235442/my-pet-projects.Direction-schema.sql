CREATE TABLE `Direction` (
  `startPlaceId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `endPlaceId` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `directionsData` json NOT NULL,
  UNIQUE KEY `Direction_startPlaceId_endPlaceId_key` (`startPlaceId`,`endPlaceId`),
  KEY `Direction_startPlaceId_idx` (`startPlaceId`),
  KEY `Direction_endPlaceId_idx` (`endPlaceId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
