
LOCK TABLES `emails` WRITE;
/*!40000 ALTER TABLE `emails` DISABLE KEYS */;
INSERT INTO `emails` VALUES (1,'a@example.com'),(2,'b@example.com');
/*!40000 ALTER TABLE `emails` ENABLE KEYS */;
UNLOCK TABLES;
