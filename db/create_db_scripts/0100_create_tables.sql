CREATE TABLE tuser (
    uuid          VARCHAR(36),
    ftime_create  TIMESTAMP DEFAULT CURRENT_timestamp,
    ftime_update  TIMESTAMP DEFAULT CURRENT_timestamp,
    id            VARCHAR(200),
    name	  VARCHAR(500),
    fdata         VARCHAR(7000)
);
CREATE UNIQUE INDEX tuser_IDX1 ON tuser (uuid);
CREATE        INDEX tuser_IDX2 ON tuser (id);

CREATE TABLE tuser_auth_vk (
    uuid          VARCHAR(36),
    uuid_tuser    VARCHAR(36),
    ftime_create  TIMESTAMP DEFAULT CURRENT_timestamp,
    ftime_update  TIMESTAMP DEFAULT CURRENT_timestamp,
    access_token  VARCHAR(1000),
    fdata         VARCHAR(7000)
);
CREATE UNIQUE INDEX tuser_auth_vk_IDX1 ON tuser_auth_vk (uuid);
CREATE        INDEX tuser_auth_vk_IDX2 ON tuser_auth_vk (uuid_tuser,ftime_create);

/************************************************************************/
SET TERM ^ ;
CREATE TRIGGER tuser_BI0 FOR tuser
ACTIVE BEFORE INSERT POSITION 0
AS
begin
  if (new.uuid is null) then begin
    new.uuid = uuid_to_char(gen_uuid());
  end
end
^

CREATE TRIGGER tuser_auth_vk_BI0 FOR tuser_auth_vk
ACTIVE BEFORE INSERT POSITION 0
AS
begin
  if (new.uuid is null) then begin
    new.uuid = uuid_to_char(gen_uuid());
  end
end
^
SET TERM ; ^
/************************************************************************/

COMMIT WORK;



