CREATE TABLE tuser (
    uuid          VARCHAR(36),
    type          INTEGER,      -- 0 - admin, 10 - spec, 100 - user  
    fam           VARCHAR(200),
    name          VARCHAR(200),
    otch 	  VARCHAR(200),
    mail          VARCHAR(200),
    phone         VARCHAR(20),
    pass          VARCHAR(100),
    street        VARCHAR(200),
    house         VARCHAR(20),
    flat          VARCHAR(20),
    info          VARCHAR(2000)
);
CREATE UNIQUE INDEX tuser_IDX1 ON tuser (uuid);
CREATE        INDEX tuser_IDX2 ON tuser (mail);


CREATE TABLE tpost (
    uuid_user     VARCHAR(36),
    uuid          VARCHAR(36),
    uuid_parent   VARCHAR(36),
    ispublic      INTEGER,
    type          VARCHAR(200),
    name          VARCHAR(500),
    text 	  VARCHAR(2000),
    data          BLOB,
    postdate      VARCHAR(40),
    postdatet     TIMESTAMP
);
CREATE UNIQUE INDEX tpost_IDX1 ON tpost (uuid);
CREATE        INDEX tpost_IDX2 ON tpost (uuid_user);
CREATE        INDEX tpost_IDX3 ON tpost (postdate,uuid_parent);


CREATE TABLE timage (
    uuid_post     VARCHAR(36),
    uuid          VARCHAR(36),
    hash          VARCHAR(200),
    title         VARCHAR(2000),
    path 	  VARCHAR(2000),
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
