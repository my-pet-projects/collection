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
