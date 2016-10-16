
ALTER TABLE timage ADD uuid_comment VARCHAR(36);
CREATE        INDEX timage_IDX4 ON timage (uuid_comment);

CREATE TABLE tcomment (
    uuid_post     VARCHAR(36),
    uuid_user     VARCHAR(36),
    uuid          VARCHAR(36),
    uuid_parent   VARCHAR(36),
    ishide        INTEGER DEFAULT 0,
    ishideuser    INTEGER DEFAULT 0,
    userdata      VARCHAR(7000),
    text 	  VARCHAR(7000),
    data          BLOB,
    upddate       VARCHAR(20),
    commentdate   VARCHAR(20),
    commentdatet  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    activecode    VARCHAR(36),
    isactive      INTEGER DEFAULT 0
);
CREATE UNIQUE INDEX tcomment_IDX1 ON tcomment (uuid);
CREATE        INDEX tcomment_IDX2 ON tcomment (uuid_user,activecode);
CREATE        INDEX tcomment_IDX3 ON tcomment (isactive);
CREATE        INDEX tcomment_IDX4 ON tcomment (uuid_post,commentdatet);

