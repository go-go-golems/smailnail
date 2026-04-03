import { useState } from "react";
import Box from "@mui/material/Box";
import {
  useExecuteQueryMutation,
  useGetPresetsQuery,
  useGetSavedQueriesQuery,
} from "../api/annotations";
import { QueryEditor } from "../components/QueryEditor";
import type { SavedQuery, QueryResult, QueryError } from "../types/annotations";

export function QueryEditorPage() {
  const [sql, setSql] = useState(
    "SELECT tag, COUNT(*) as count\nFROM annotations\nGROUP BY tag\nORDER BY count DESC;",
  );
  const [result, setResult] = useState<QueryResult | null>(null);
  const [error, setError] = useState<QueryError | null>(null);

  const { data: presets = [] } = useGetPresetsQuery();
  const { data: savedQueries = [] } = useGetSavedQueriesQuery();
  const [executeQuery, { isLoading }] = useExecuteQueryMutation();

  const handleExecute = async (querySql: string) => {
    setError(null);
    try {
      const res = await executeQuery({ sql: querySql }).unwrap();
      setResult(res);
    } catch (err: unknown) {
      const message =
        err && typeof err === "object" && "data" in err
          ? String((err as { data: { message: string } }).data.message)
          : "Query execution failed";
      setError({ message });
      setResult(null);
    }
  };

  const handleSelectQuery = (query: SavedQuery) => {
    setSql(query.sql);
    setResult(null);
    setError(null);
  };

  return (
    <Box
      data-widget="query-editor-page"
      sx={{ height: "calc(100vh - 0px)", display: "flex" }}
    >
      <QueryEditor
        sql={sql}
        onSqlChange={setSql}
        onExecute={(s) => void handleExecute(s)}
        onSelectQuery={handleSelectQuery}
        presets={presets}
        savedQueries={savedQueries}
        result={result}
        error={error}
        isLoading={isLoading}
      />
    </Box>
  );
}
