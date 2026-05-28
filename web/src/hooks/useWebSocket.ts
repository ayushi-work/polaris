import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { wsClient } from "@/api/websocket";
import type { WSEvent, Incident, Remediation, RCAResult, ChaosScenario } from "@/types";

export function useWebSocket() {
  const queryClient = useQueryClient();

  useEffect(() => {
    wsClient.connect();

    const unsubs = [
      wsClient.on("incident.created", (e: WSEvent) => {
        queryClient.invalidateQueries({ queryKey: ["incidents"] });
        queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      }),
      wsClient.on("incident.updated", (e: WSEvent) => {
        const inc = e.payload as Incident;
        queryClient.setQueryData(["incident", inc.id], inc);
        queryClient.invalidateQueries({ queryKey: ["incidents"] });
        queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      }),
      wsClient.on("remediation.started", (e: WSEvent) => {
        queryClient.invalidateQueries({ queryKey: ["remediations"] });
      }),
      wsClient.on("remediation.completed", (e: WSEvent) => {
        queryClient.invalidateQueries({ queryKey: ["remediations"] });
      }),
      wsClient.on("rca.completed", (e: WSEvent) => {
        queryClient.invalidateQueries({ queryKey: ["rca"] });
      }),
      wsClient.on("chaos.executing", (e: WSEvent) => {
        queryClient.invalidateQueries({ queryKey: ["scenarios"] });
      }),
      wsClient.on("chaos.completed", (e: WSEvent) => {
        queryClient.invalidateQueries({ queryKey: ["scenarios"] });
      }),
    ];

    return () => {
      unsubs.forEach((u) => u());
      wsClient.disconnect();
    };
  }, [queryClient]);
}
