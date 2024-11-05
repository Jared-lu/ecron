CREATE DATABASE IF NOT EXISTS ecron;

CREATE TABLE IF NOT EXISTS `ecron.task_info`
(
    id                BIGINT AUTO_INCREMENT PRIMARY KEY ,
    name              VARCHAR(128)  NOT NULL COMMENT '任务名称',
    type              VARCHAR(32)   NOT NULL COMMENT '任务类型',
    cron              VARCHAR(32)   NOT NULL COMMENT 'cron表达式',
    executor          VARCHAR(32)  NOT NULL COMMENT '执行器名称',
    owner             VARCHAR(64)   NOT NULL COMMENT '用于实现乐观锁',
    status            TINYINT NOT NULL DEFAULT 1 COMMENT '0-无效，1-有效',
    cfg               TEXT          NOT NULL COMMENT '任务配置',
    next_exec_time    BIGINT COMMENT '下一次执行时间',
    ctime       BIGINT        NOT NULL ,
    utime      BIGINT        NOT NULL ,
    INDEX idx_status_next_exec_time(status, next_exec_time),
    INDEX idx_status_utime(status, utime)
) COMMENT '任务信息';


CREATE TABLE IF NOT EXISTS  `ecron.task_info`
(
    id          BIGINT AUTO_INCREMENT PRIMARY KEY ,
    tid         BIGINT NOT NULL COMMENT '任务id',
    status      TINYINT COMMENT '执行状态，0-未知，1-运行中，2-成功，3-失败，4-超时，5-主动取消',
    progress    INT COMMENT '执行进度，取值0-100',
    ctime       BIGINT        NOT NULL ,
    utime       bigint        NOT NULL,
    UNIQUE idx_tid(tid)
) comment '任务执行情况';
