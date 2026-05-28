import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchScenarios, executeScenario, createScenario } from "@/api/chaos";

export function useChaosScenarios() {
  return useQuery({
    queryKey: ["scenarios"],
    queryFn: fetchScenarios,
  });
}

export function useExecuteScenario() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: executeScenario,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["scenarios"] }),
  });
}

export function useCreateScenario() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createScenario,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["scenarios"] }),
  });
}
