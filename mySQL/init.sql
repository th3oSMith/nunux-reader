-- Création des Tables

/* Désactivation des contraintes */
SET foreign_key_checks = 0;

/* Table contenant les utilisateurs */
DROP TABLE IF EXISTS user;
CREATE TABLE user (
                      id INT NOT NULL AUTO_INCREMENT,
                      username VARCHAR(50) NOT NULL,
                      password VARCHAR(255) NOT NULL,
                      saved_timeline_id INT(4) NOT NULL,
                      PRIMARY KEY ( id ),
                      FOREIGN KEY (saved_timeline_id) REFERENCES timeline(id)
                      ) ENGINE=INNODB;

DROP TABLE IF EXISTS feed;
CREATE TABLE feed (
                      id INT NOT NULL AUTO_INCREMENT,
                      nickname VARCHAR(255) NOT NULL,
                      title VARCHAR(255) NOT NULL,
                      description TEXT NOT NULL,
                      link VARCHAR(255) NOT NULL,
                      updateUrl VARCHAR(255) NOT NULL,
                      refresh DATETIME NOT NULL,
                      unread INT NOT NULL,
                      insecure INT(1) NOT NULL,
                      username VARCHAR(50) NOT NULL,
                      password VARCHAR(50) NOT NULL,
                      PRIMARY KEY (id)
                      ) ENGINE=INNODB;

DROP TABLE IF EXISTS article;
CREATE TABLE article (
                      id INT NOT NULL AUTO_INCREMENT,
                      date DATETIME NOT NULL,
                      description LONGTEXT NOT NULL,
                      link VARCHAR(255) NOT NULL,
                      pubdate DATETIME NOT NULL,
                      title VARCHAR(255) NOT NULL,
                      feed_id INT(4) NOT NULL,
                      PRIMARY KEY (id),
                      FOREIGN KEY (feed_id) REFERENCES feed(id)
                      ) ENGINE=INNODB;


DROP TABLE IF EXISTS timeline;
CREATE TABLE timeline (
                      id INT NOT NULL AUTO_INCREMENT,
                      timeline VARCHAR(50) NOT NULL,
                      title VARCHAR(255) NOT NULL,
                      size SMALLINT NOT NULL,
                      feed_id INT(4),
                      user_id INT,
                      PRIMARY KEY (id),
                      FOREIGN KEY (feed_id) REFERENCES feed(id),
                      FOREIGN KEY (user_id) REFERENCES user(id)
                      ) ENGINE=INNODB;

DROP TABLE IF EXISTS article_timelines;
CREATE TABLE article_timelines (
                      article_id INT(4) NOT NULL,
                      timeline_id INT(4) NOT NULL,
                      delete_date DATE,
                      FOREIGN KEY (article_id) REFERENCES article(id),
                      FOREIGN KEY (timeline_id) REFERENCES timeline(id)
                      ) ENGINE=INNODB;

/* Réactivation des contraintes */
SET foreign_key_checks = 1;