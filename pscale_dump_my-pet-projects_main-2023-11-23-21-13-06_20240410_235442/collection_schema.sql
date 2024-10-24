CREATE TABLE `countries` (
  `cca2` varchar(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cca3` varchar(3) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ccn3` varchar(3) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name_common` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name_official` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `region` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `subregion` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `flag_url` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`cca2`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `cities` (
  `oldIdForDelete` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `id` bigint NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `alternate_names` text COLLATE utf8mb4_unicode_ci,
  `country_code` varchar(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `admin1_code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `admin2_code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `admin3_code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `admin4_code` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `modification_date` date NOT NULL,
  `population` int DEFAULT NULL,
  `latitude` double NOT NULL,
  `longitude` double NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_cities_countries_code_idx` (`country_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `breweries` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `website` varchar(100) DEFAULT NULL,
  `geo_id` bigint NOT NULL,
  `old_id` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_name_geo_id` (`name`,`geo_id`),
  KEY `fk_geo_id` (`geo_id`),
  CONSTRAINT `fk_geo_id` FOREIGN KEY (`geo_id`) REFERENCES `cities` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=683 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `beer_styles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  PRIMARY KEY (`id`)
);



CREATE TABLE `beers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `brand` varchar(100) NOT NULL,
  `type` varchar(100) DEFAULT NULL,
  `style_id` int NOT NULL,
  `brewery_id` bigint DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '0',
  `created_at` datetime NOT NULL,
  `updated_at` datetime DEFAULT NULL,
  `old_image_ids` varchar(200) NULL,
  PRIMARY KEY (`id`),
  KEY `fk_brewery_id` (`brewery_id`),
  CONSTRAINT `fk_brewery_id` FOREIGN KEY (`brewery_id`) REFERENCES `breweries` (`id`)
  CONSTRAINT `fk_style_id` FOREIGN KEY (`style_id`) REFERENCES `beer_styles` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2168 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;




CREATE TABLE `beer_medias` (
  `id` int NOT NULL AUTO_INCREMENT,
  `beer_id` int NOT NULL,
  `media_id` int NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_beer_id` (`beer_id`),
  CONSTRAINT `fk_beer_id` FOREIGN KEY (`beer_id`) REFERENCES `beers` (`id`)
  CONSTRAINT `fk_media_id` FOREIGN KEY (`media_id`) REFERENCES `media_items` (`id`)
)

CREATE TABLE `media_items` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `filename` varchar(100) NOT NULL,
  `type` int NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE `media_types` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  PRIMARY KEY (`id`)
)

INSERT INTO `media_types` VALUES 
(1,'Bottle'),
(2,'Label'),
(3,'Crown Cap'),
(4,'Twist-off Cap'),
(5,'Pull-off Cap'),
(6,'Ceramic Cap');
