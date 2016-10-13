CREATE TABLE tuser (
    uuid          VARCHAR(36),
    type          INTEGER DEFAULT 0,      -- 100 - admin, 50 - moderator, 10 - spec, 0 - user  
    fam           VARCHAR(200),
    name          VARCHAR(200),
    pat 	  VARCHAR(200),
    email         VARCHAR(200),
    phone         VARCHAR(20),
    pass          VARCHAR(100),
    street        VARCHAR(200),
    house         VARCHAR(200),
    flat          VARCHAR(200),
    info          VARCHAR(5000),
    upddate       VARCHAR(20),
    regdate       VARCHAR(20),
    regdatet      TIMESTAMP,
    lastvisit     VARCHAR(20),
    isactive      INTEGER DEFAULT 0,
    activecode    VARCHAR(36),
    activecodepass  VARCHAR(36),   /* код пользователя который обновляется при каждой аторизации на сайте
				      поэтому изменения данных пользователя возможны только для последнего вошедшего под его учеткой 
                                   */
    istemp        INTEGER DEFAULT 0
);
CREATE UNIQUE INDEX tuser_IDX1 ON tuser (uuid,istemp);
CREATE UNIQUE INDEX tuser_IDX2 ON tuser (email,istemp);

CREATE TABLE tuser_type (
    type          INTEGER DEFAULT 0,      -- 100 - admin, 50 - moderator, 10 - spec, 0 - user  
    name          VARCHAR(200)
);
CREATE UNIQUE INDEX tuser_type_IDX1 ON tuser_type (type);
DELETE FROM tuser_type WHERE 1=1;
INSERT INTO tuser_type(type,name) VALUES(100,'Администратор');
INSERT INTO tuser_type(type,name) VALUES(50,'Модератор');
INSERT INTO tuser_type(type,name) VALUES(10,'Специалист');
INSERT INTO tuser_type(type,name) VALUES(0,'');
COMMIT;


CREATE TABLE tpost (
    uuid_user     VARCHAR(36),
    uuid          VARCHAR(36),
    uuid_parent   VARCHAR(36),
    ishide        INTEGER DEFAULT 0,
    ishideuser    INTEGER DEFAULT 0,
    type          VARCHAR(200),
    userdata      VARCHAR(7000),
    text 	  VARCHAR(7000),
    data          BLOB,
    upddate       VARCHAR(20),
    postdate      VARCHAR(20),
    postdatet     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    activecode    VARCHAR(36),
    isactive      INTEGER DEFAULT 0,
    isstartwork   INTEGER DEFAULT 0  
);
CREATE UNIQUE INDEX tpost_IDX1 ON tpost (uuid);
CREATE        INDEX tpost_IDX2 ON tpost (uuid_user,activecode);
CREATE        INDEX tpost_IDX3 ON tpost (isactive,postdatet);
CREATE        INDEX tpost_IDX4 ON tpost (isstartwork,postdatet);

CREATE TABLE timage (
    uuid_post     VARCHAR(36),
    uuid          VARCHAR(36),
    hash          VARCHAR(200),
    title         VARCHAR(2000),
    path 	  VARCHAR(200),
    pathmin	  VARCHAR(200),
    imgdate       VARCHAR(40),
    imgdatet      TIMESTAMP
);
CREATE UNIQUE INDEX timage_IDX1 ON timage (uuid);
CREATE        INDEX timage_IDX2 ON timage (uuid_post);
CREATE        INDEX timage_IDX3 ON timage (hash);


CREATE TABLE tfile (
    uuid_post     VARCHAR(36),
    uuid          VARCHAR(36),
    hash          VARCHAR(200),
    title         VARCHAR(2000),
    path 	  VARCHAR(2000),
    imgdate       VARCHAR(40),
    imgdatet      TIMESTAMP
);
CREATE UNIQUE INDEX tfile_IDX1 ON tfile (uuid);
CREATE        INDEX tfile_IDX2 ON tfile (uuid_post);
CREATE        INDEX tfile_IDX3 ON tfile (hash);


COMMIT WORK;
