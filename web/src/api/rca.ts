import api from "./client";
import type { RCAResult } from "@/types";

export async function fetchRCA(incidentId: string): Promise<RCAResult> {
  const { data } = await api.get(`/analysis/${incidentId}`);
  return data;
}

export async function triggerRCA(incidentId: string): Promise<{ status: string }> {
  const { data } = await api.post(`/analysis/${incidentId}`);
  return data;
}
