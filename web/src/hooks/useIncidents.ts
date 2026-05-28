import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchIncidents,
  fetchIncident,
  createIncident,
  acknowledgeIncident,
  resolveIncident,
} from "@/api/incidents";

export function useIncidents(params?: Record<string, string>) {
  return useQuery({
    queryKey: ["incidents", params],
    queryFn: () => fetchIncidents(params),
    refetchInterval: 15000,
  });
}

export function useIncident(id: string) {
  return useQuery({
    queryKey: ["incident", id],
    queryFn: () => fetchIncident(id),
    enabled: !!id,
  });
}

export function useCreateIncident() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createIncident,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["incidents"] }),
  });
}

export function useAcknowledgeIncident() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: acknowledgeIncident,
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: ["incidents"] });
      qc.invalidateQueries({ queryKey: ["incident", id] });
    },
  });
}

export function useResolveIncident() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: resolveIncident,
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: ["incidents"] });
      qc.invalidateQueries({ queryKey: ["incident", id] });
    },
  });
}
