SET SQL DIALECT 3;


CREATE GENERATOR ID_WORD;
SET GENERATOR ID_WORD TO 1;


SET TERM ^ ; 
CREATE PROCEDURE GET_WORD_CNT (
    WORD VARCHAR(100))
RETURNS (
    UUID_POST VARCHAR(36),
    CNT BIGINT)
AS
BEGIN
  SUSPEND;
END^
SET TERM ; ^


CREATE TABLE TPOST (
    NAME               VARCHAR(1000) CHARACTER SET WIN1251,
    TAGS               VARCHAR(4000) CHARACTER SET WIN1251,
    TEXT               VARCHAR(16000) CHARACTER SET WIN1251,
    PREVIEW            VARCHAR(2000) CHARACTER SET WIN1251,
    UUID_USER_CREATE   VARCHAR(36) CHARACTER SET WIN1251,
    UUID_USER          VARCHAR(36) CHARACTER SET WIN1251,
    DATE_MODIFY        TIMESTAMP DEFAULT current_timestamp,
    DATE_CREATE        TIMESTAMP,
    UUID               VARCHAR(36) CHARACTER SET WIN1251,
    UUID_USER_PUBLISH  VARCHAR(36) CHARACTER SET WIN1251,
    DATE_PUBLISH       TIMESTAMP,
    EDIT_TYPE          VARCHAR(10) CHARACTER SET WIN1251
);
/*******************
UUID_USER_CREATE - пользователь который создал запись
UUID_USER - последний пользователь который внес изменения в пост
UUID_USER_PUBLISH - пользователь который подтвердил изменения и разрешил публикацию поста
********************/

CREATE TABLE TPOST_LOG (
    OPER               VARCHAR(10) CHARACTER SET WIN1251,
    NAME               VARCHAR(1000) CHARACTER SET WIN1251,
    TAGS               VARCHAR(4000) CHARACTER SET WIN1251,
    TEXT               VARCHAR(16000) CHARACTER SET WIN1251,
    PREVIEW            VARCHAR(2000) CHARACTER SET WIN1251,
    UUID_USER_CREATE   VARCHAR(36) CHARACTER SET WIN1251,
    UUID_USER          VARCHAR(36) CHARACTER SET WIN1251,
    DATE_MODIFY        TIMESTAMP,
    DATE_CREATE        TIMESTAMP DEFAULT current_timestamp,
    UUID               VARCHAR(36) CHARACTER SET WIN1251,
    UUID_USER_PUBLISH  VARCHAR(36),
    DATE_PUBLISH       TIMESTAMP,
    EDIT_TYPE          VARCHAR(10)
);

CREATE TABLE TWORD (
    ID    BIGINT,
    WORD  VARCHAR(100) CHARACTER SET WIN1251
);

CREATE TABLE TWORD_POST (
    UUID_POST  VARCHAR(36),
    ID_WORD    BIGINT,
    CNT        BIGINT
);

CREATE UNIQUE INDEX IDX_TPOST_1 ON TPOST (UUID, EDIT_TYPE, UUID_USER, UUID_USER_PUBLISH);
CREATE INDEX IDX_TPOST_2 ON TPOST (EDIT_TYPE, UUID);
CREATE INDEX IDX_TPOST_3 ON TPOST (UUID_USER);
CREATE DESCENDING INDEX TPOST_IDX1 ON TPOST (EDIT_TYPE, DATE_CREATE);
CREATE INDEX IDX_TPOST_LOG_1 ON TPOST_LOG (UUID, DATE_CREATE);
CREATE INDEX IDX_TPOST_LOG_2 ON TPOST_LOG (UUID_USER);
CREATE INDEX IDX_TPOST_LOG_3 ON TPOST_LOG (OPER, DATE_CREATE);
CREATE UNIQUE INDEX IDX_TWORD_1 ON TWORD (ID);
CREATE UNIQUE INDEX IDX_TWORD_2 ON TWORD (WORD);
CREATE INDEX IDX_TWORD_POST_1 ON TWORD_POST (UUID_POST, CNT);
CREATE INDEX IDX_TWORD_POST_2 ON TWORD_POST (ID_WORD, CNT);


SET TERM ^ ;


CREATE TRIGGER TPOST_BD0 FOR TPOST
ACTIVE BEFORE DELETE POSITION 0
AS
begin
    INSERT INTO TPOST_LOG(oper,name,tags,text,preview,uuid_user_create,uuid_user,date_modify,uuid,uuid_user_publish,date_publish,edit_type)
    VALUES('delete',old.name,old.tags,old.text,old.preview,old.uuid_user_create,old.uuid_user,old.date_modify,old.uuid,old.uuid_user_publish,old.date_publish,old.edit_type);
end
^

CREATE TRIGGER TPOST_BI0 FOR TPOST
ACTIVE BEFORE INSERT POSITION 0
AS
begin
  if (new.date_create is null) then new.date_create = current_timestamp;
  if (new.date_modify is null) then new.date_modify = current_timestamp;
  if (new.uuid is null) then new.uuid = uuid_to_char(gen_uuid());
end
^

CREATE TRIGGER TPOST_BU0 FOR TPOST
ACTIVE BEFORE UPDATE POSITION 0
AS
begin
    new.date_modify = current_timestamp;
    new.date_create = old.date_create;
    new.uuid = old.uuid;
    INSERT INTO TPOST_LOG(oper,name,tags,text,preview,uuid_user_create,uuid_user,date_modify,uuid,uuid_user_publish,date_publish,edit_type)
    VALUES('update',old.name,old.tags,old.text,old.preview,old.uuid_user_create,old.uuid_user,old.date_modify,old.uuid,old.uuid_user_publish,old.date_publish,old.edit_type);
end
^


SET TERM ; ^



SET TERM ^ ;
ALTER PROCEDURE GET_WORD_CNT (
    WORD VARCHAR(100))
RETURNS (
    UUID_POST VARCHAR(36),
    CNT BIGINT)
AS
 BEGIN 
  for
    SELECT wp.uuid_post,SUM(wp.cnt)
    FROM tword_post wp
      INNER JOIN tword w ON w.id=wp.id_word
    WHERE w.word LIKE :word
    GROUP BY wp.uuid_post
    INTO :uuid_post,:cnt
  do begin
    suspend;
  end
END^
SET TERM ; ^
