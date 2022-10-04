-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

CREATE TABLE IF NOT EXISTS device_nhtsa_recalls
(
    id character(27) COLLATE pg_catalog."default" NOT NULL,
    device_definition_id character(27) COLLATE pg_catalog."default" NOT NULL,
    data_record_id integer COLLATE pg_catalog."default" NOT NULL,
    data_campno character varying(12) COLLATE pg_catalog."default" NOT NULL,
    data_maketxt character varying(25) COLLATE pg_catalog."default" NOT NULL,
    data_modeltxt character varying(256) COLLATE pg_catalog."default" NOT NULL,
    data_yeartxt integer COLLATE pg_catalog."default" NOT NULL,
    data_mfgcampno character varying(20) COLLATE pg_catalog."default" NOT NULL,
    data_compname character varying(256) COLLATE pg_catalog."default" NOT NULL,
    data_mfgname character varying(40) COLLATE pg_catalog."default" NOT NULL,
    data_bgman date COLLATE pg_catalog."default" NULL,
    data_endman date COLLATE pg_catalog."default" NULL,
    data_rcltypecd character varying(4) COLLATE pg_catalog."default" NOT NULL,
    data_potaff integer COLLATE pg_catalog."default" NULL,
    data_odate date COLLATE pg_catalog."default" NULL,
    data_influenced_by character varying(4) COLLATE pg_catalog."default" NOT NULL,
    data_mfgtxt character varying(40) COLLATE pg_catalog."default" NOT NULL,
    data_rcdate date COLLATE pg_catalog."default" NOT NULL,
    data_datea date COLLATE pg_catalog."default" NOT NULL,
    data_rpno character varying(3) COLLATE pg_catalog."default" NOT NULL,
    data_fmvss character varying(10) COLLATE pg_catalog."default" NOT NULL,
    data_desc_defect character varying(2000) COLLATE pg_catalog."default" NOT NULL,
    data_conequence_defect character varying(2000) COLLATE pg_catalog."default" NOT NULL,
    data_corrective_action character varying(2000) COLLATE pg_catalog."default" NOT NULL,
    data_notes character varying(2000) COLLATE pg_catalog."default" NOT NULL,
    data_rcl_cmpt_id character(27) COLLATE pg_catalog."default" NOT NULL,
    data_mfr_comp_name character varying(50) COLLATE pg_catalog."default" NOT NULL,
    data_mfr_comp_desc character varying(200) COLLATE pg_catalog."default" NOT NULL,
    data_mfr_comp_ptno character varying(100) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata jsonb COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT device_nhtsa_recalls_pkey PRIMARY KEY (id),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id)
        REFERENCES device_definitions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

/*
NHTSA Recall Data - Flat File Information
https://static.nhtsa.gov/odi/ffdd/rcl/RCL.txt
*/
COMMENT ON COLUMN device_nhtsa_recalls.data_record_id         IS '1. RUNNING SEQUENCE NUMBER, WHICH UNIQUELY IDENTIFIES THE RECORD';
COMMENT ON COLUMN device_nhtsa_recalls.data_campno            IS '2. NHTSA CAMPAIGN NUMBER';
COMMENT ON COLUMN device_nhtsa_recalls.data_maketxt           IS '3. VEHICLE/EQUIPMENT MAKE';
COMMENT ON COLUMN device_nhtsa_recalls.data_modeltxt          IS '4. VEHICLE/EQUIPMENT MODEL';
COMMENT ON COLUMN device_nhtsa_recalls.data_yeartxt           IS '5. MODEL YEAR, 9999 IF UNKNOWN or N/A';
COMMENT ON COLUMN device_nhtsa_recalls.data_mfgcampno         IS '6. MFR CAMPAIGN NUMBER';
COMMENT ON COLUMN device_nhtsa_recalls.data_compname          IS '7. COMPONENT DESCRIPTION';
COMMENT ON COLUMN device_nhtsa_recalls.data_mfgname           IS '8. MANUFACTURER THAT FILED DEFECT/NONCOMPLIANCE REPORT';
COMMENT ON COLUMN device_nhtsa_recalls.data_bgman             IS '9. BEGIN DATE OF MANUFACTURING';
COMMENT ON COLUMN device_nhtsa_recalls.data_endman            IS '10. END DATE OF MANUFACTURING';
COMMENT ON COLUMN device_nhtsa_recalls.data_rcltypecd         IS '11. VEHICLE, EQUIPMENT OR TIRE REPORT';
COMMENT ON COLUMN device_nhtsa_recalls.data_potaff            IS '12. POTENTIAL NUMBER OF UNITS AFFECTED';
COMMENT ON COLUMN device_nhtsa_recalls.data_odate             IS '13. DATE OWNER NOTIFIED BY MFR';
COMMENT ON COLUMN device_nhtsa_recalls.data_influenced_by     IS '14. RECALL INITIATOR (MFR/OVSC/ODI)';
COMMENT ON COLUMN device_nhtsa_recalls.data_mfgtxt            IS '15. MANUFACTURERS OF RECALLED VEHICLES/PRODUCTS';
COMMENT ON COLUMN device_nhtsa_recalls.data_rcdate            IS '16. REPORT RECEIVED DATE';
COMMENT ON COLUMN device_nhtsa_recalls.data_datea             IS '17. RECORD CREATION DATE';
COMMENT ON COLUMN device_nhtsa_recalls.data_rpno              IS '18. REGULATION PART NUMBER';
COMMENT ON COLUMN device_nhtsa_recalls.data_fmvss             IS '19. FEDERAL MOTOR VEHICLE SAFETY STANDARD NUMBER';
COMMENT ON COLUMN device_nhtsa_recalls.data_desc_defect       IS '20. DEFECT SUMMARY';
COMMENT ON COLUMN device_nhtsa_recalls.data_conequence_defect IS '21. CONSEQUENCE SUMMARY';
COMMENT ON COLUMN device_nhtsa_recalls.data_corrective_action IS '22. CORRECTIVE SUMMARY';
COMMENT ON COLUMN device_nhtsa_recalls.data_notes             IS '23. RECALL NOTES';
COMMENT ON COLUMN device_nhtsa_recalls.data_rcl_cmpt_id       IS '24. NUMBER THAT UNIQUELY IDENTIFIES A RECALLED COMPONENT';
COMMENT ON COLUMN device_nhtsa_recalls.data_mfr_comp_name     IS '25. MANUFACTURER-SUPPLIED COMPONENT NAME';
COMMENT ON COLUMN device_nhtsa_recalls.data_mfr_comp_desc     IS '26. MANUFACTURER-SUPPLIED COMPONENT DESCRIPTION';
COMMENT ON COLUMN device_nhtsa_recalls.data_mfr_comp_ptno     IS '27. MANUFACTURER-SUPPLIED COMPONENT PART NUMBER';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

DROP TABLE device_nhtsa_recalls;

-- +goose StatementEnd
