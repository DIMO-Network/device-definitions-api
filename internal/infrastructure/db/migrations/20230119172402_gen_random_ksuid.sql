-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO device_definitions_api,public;
/**
 * Returns a Segment's KSUID.
 *
 * Reference implementation: https://github.com/segmentio/ksuid
 * Also read: https://segment.com/blog/a-brief-history-of-the-uuid/
 * https://gist.github.com/fabiolimace/5e7923803566beefaf3c716d1343ae27
 * LICENSE: https://creativecommons.org/licenses/by-sa/4.0/ (Gist copied from stackoverflow)
 */
create or replace function gen_random_ksuid() returns text as $$
declare
    v_time timestamp with time zone := null;
    v_seconds numeric := null;
    v_payload bytea := null;
    v_numeric numeric := null;
    v_base62 text := '';
    v_epoch numeric = 1400000000; -- 2014-05-13T16:53:20Z
    v_alphabet char array[62] := array[
        '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
        'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
        'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
        'U', 'V', 'W', 'X', 'Y', 'Z',
        'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
        'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
        'u', 'v', 'w', 'x', 'y', 'z'];
    i integer := 0;
begin

    -- Get the current time
    v_time := clock_timestamp();

    -- Extract seconds from the current time and apply epoch
    v_seconds := EXTRACT(EPOCH FROM v_time) - v_epoch;

    -- Generate a numeric value from the seconds
    v_numeric := v_seconds * pow(2::numeric, 128);

    -- Generate a pseudo-random payload
    v_payload := decode(md5(v_time::text || random()::text || random()::text), 'hex');

    -------------------------------------------------------------------
    -- FOR TEST: the expected result is '0ujtsYcgvSTl8PAuAdqWYSMnLOv'
    -------------------------------------------------------------------
    -- v_numeric := 107608047 * pow(2::numeric, 128);
    -- v_payload := decode('B5A1CD34B5F99D1154FB6853345C9735', 'hex');

    -- Add the payload to the numeric value
    while i < 16 loop
            i := i + 1;
            v_numeric := v_numeric + (get_byte(v_payload, i - 1) * pow(2::numeric, (16 - i) * 8));
        end loop;

    -- Encode the numeric value to base62
    while v_numeric <> 0 loop
            v_base62 := v_base62 || v_alphabet[mod(v_numeric, 62) + 1];
            v_numeric := div(v_numeric, 62);
        end loop;
    v_base62 := reverse(v_base62);
    v_base62 := lpad(v_base62, 27, '0');

    return v_base62;

end $$ language plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api,public;

DROP FUNCTION gen_random_ksuid();
-- +goose StatementEnd
