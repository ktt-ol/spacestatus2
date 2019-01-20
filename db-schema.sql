
CREATE TABLE IF NOT EXISTS `devices` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `devices` int(11) NOT NULL DEFAULT '0',
  `people` int(11) NOT NULL DEFAULT '0',
  `ts` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `spacestate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `state` varchar(50) NOT NULL,
  `until` datetime DEFAULT NULL,
  `lastupdate` datetime DEFAULT NULL,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `place` varchar(20) NOT NULL DEFAULT 'space',
  PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1;
