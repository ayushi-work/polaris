import api from "./client";
import type { ChaosScenario } from "@/types";

export async function fetchScenarios(): Promise<ChaosScenario[]> {
  const { data } = await api.get("/chaos/scenarios");
  return data;
}

export async function executeScenario(id: string): Promise<{ status: string }> {
  const { data } = await api.post(`/chaos/scenarios/${id}/execute`);
  return data;
}

export async function createScenario(scenario: Partial<ChaosScenario>): Promise<ChaosScenario> {
  const { data } = await api.post("/chaos/scenarios", scenario);
  return data;
}
