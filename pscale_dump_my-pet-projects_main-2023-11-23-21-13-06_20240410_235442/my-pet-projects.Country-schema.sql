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
