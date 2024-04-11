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
