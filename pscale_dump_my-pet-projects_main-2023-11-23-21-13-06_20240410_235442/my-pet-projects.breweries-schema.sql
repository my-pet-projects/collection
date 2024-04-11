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
