create table filetosystemmap (fileid bigint not null, systemid bigint not null, filepath text not null, firstseen bigint not null, lastseen bigint not null, PRIMARY KEY(fileid, systemid));


-- Must create aggregate First function as defined: https://wiki.postgresql.org/wiki/First/last_(aggregate)

CREATE OR REPLACE FUNCTION public.first_agg ( anyelement, anyelement )
RETURNS anyelement LANGUAGE sql IMMUTABLE STRICT AS $$
        SELECT $1;
$$;
 
-- And then wrap an aggregate around it
CREATE AGGREGATE public.first (
        sfunc    = public.first_agg,
        basetype = anyelement,
        stype    = anyelement
);
