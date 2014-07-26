-- Création des Tables

/* Table contenant les utilisateurs */
DROP TABLE IF EXISTS user;
CREATE TABLE user (
                      id INT NOT NULL AUTO_INCREMENT,
                      username VARCHAR(50) NOT NULL,
                      password VARCHAR(50) NOT NULL,
                      PRIMARY KEY ( id )
                      ) ENGINE=INNODB;

DROP TABLE IF EXISTS feed;
CREATE TABLE feed (
                      id INT NOT NULL AUTO_INCREMENT,
                      title VARCHAR(255) NOT NULL,
                      xmlurl VARCHAR(255) NOT NULL,
                      status VARCHAR(50) NOT NULL,
                      lastModified DATETIME NOT NULL,
                      expires DATETIME NOT NULL,
                      etag VARCHAR(255) NOT NULL,
                      updateTime DATETIME NOT NULL,
                      errCount SMALLINT NOT NULL,
                      description TEXT NOT NULL,
                      link VARCHAR(255) NOT NULL,
                      hub VARCHAR(255) NOT NULL,
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
                      PRIMARY KEY (id),
                      FOREIGN KEY (feed_id) REFERENCES feed(id)
                      ) ENGINE=INNODB;

DROP TABLE IF EXISTS user_timelines;
CREATE TABLE user_timelines (
                      user_id INT(4) NOT NULL,
                      timeline_id INT(4) NOT NULL,
                      FOREIGN KEY (user_id) REFERENCES user(id),
                      FOREIGN KEY (timeline_id) REFERENCES timeline(id)
                      ) ENGINE=INNODB;
