SET SQL DIALECT 3;


CREATE TABLE TPOST (
    NAME              VARCHAR(1000) DEFAULT '',
    TAGS              VARCHAR(4000),
    TEXT              VARCHAR(16000),
    PREVIEW           VARCHAR(2000),
    UUID_USER         VARCHAR(36),
    DATE_MODIFY       TIMESTAMP DEFAULT current_timestamp,
    DATE_CREATE       TIMESTAMP,
    UUID              VARCHAR(36)
);

CREATE TABLE TPOST_LOG (
    OPER              VARCHAR(10),
    NAME              VARCHAR(1000),
    TAGS              VARCHAR(4000),
    TEXT              VARCHAR(16000),
    PREVIEW           VARCHAR(2000),
    UUID_USER         VARCHAR(36),
    DATE_MODIFY       TIMESTAMP,
    DATE_CREATE       TIMESTAMP DEFAULT current_timestamp,
    UUID              VARCHAR(36)
);




/******************************************************************************/
/***                                Indices                                 ***/
/******************************************************************************/

CREATE DESCENDING INDEX TPOST_IDX1   ON TPOST (DATE_CREATE);
CREATE UNIQUE     INDEX IDX_TPOST_1  ON TPOST (UUID);
CREATE            INDEX IDX_TPOST_2  ON TPOST (UUID_USER);

CREATE       INDEX IDX_TPOST_LOG_1   ON TPOST_LOG (UUID,DATE_CREATE);
CREATE       INDEX IDX_TPOST_LOG_2   ON TPOST_LOG (UUID_USER);
CREATE       INDEX IDX_TPOST_LOG_3   ON TPOST_LOG (OPER,DATE_CREATE);

/******************************************************************************/
/***                                Triggers                                ***/
/******************************************************************************/


SET TERM ^ ;



/******************************************************************************/
/***                          Triggers for tables                           ***/
/******************************************************************************/



/* Trigger: TPOST_BI0 */
CREATE TRIGGER TPOST_BI0 FOR TPOST
ACTIVE BEFORE INSERT POSITION 0
AS
begin
  if (new.date_create is null) then new.date_create = current_timestamp;
  if (new.date_modify is null) then new.date_modify = current_timestamp;
  if (new.uuid is null) then new.uuid = uuid_to_char(gen_uuid());
end
^

/* Trigger: TPOST_BU0 */
CREATE TRIGGER TPOST_BU0 FOR TPOST
ACTIVE BEFORE UPDATE POSITION 0
AS
begin
    new.date_modify = current_timestamp;
    new.date_create = old.date_create;
    new.uuid = old.uuid;

    INSERT INTO TPOST_LOG(OPER,NAME,TAGS,TEXT,PREVIEW,UUID_USER,DATE_MODIFY,UUID)
    VALUES('update',old.name,old.tags,old.text,old.preview,old.uuid_user,old.date_modify,old.uuid);
end
^

/* лог триггер на удаление записей */
CREATE TRIGGER TPOST_BD0 FOR TPOST
ACTIVE BEFORE DELETE POSITION 0
AS
begin
    INSERT INTO TPOST_LOG(OPER,NAME,TAGS,TEXT,PREVIEW,UUID_USER,DATE_MODIFY,UUID)
    VALUES('delete',old.name,old.tags,old.text,old.preview,old.uuid_user,old.date_modify,old.uuid);
end
^

SET TERM ; ^





CREATE TABLE TWORD (
    ID           BIGINT,
    WORD         VARCHAR(100)
);
CREATE UNIQUE INDEX IDX_TWORD_1 ON TWORD (ID);
CREATE UNIQUE INDEX IDX_TWORD_2 ON TWORD (WORD);


CREATE TABLE TWORD_POST (
    UUID_POST    VARCHAR(36),
    ID_WORD      BIGINT,
    CNT          BIGINT
);
CREATE INDEX IDX_TWORD_POST_1 ON TWORD_POST (UUID_POST, CNT);
CREATE INDEX IDX_TWORD_POST_2 ON TWORD_POST (ID_WORD, CNT);





