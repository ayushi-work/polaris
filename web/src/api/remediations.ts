import api from "./client";
import type { Remediation } from "@/types";

export async function fetchRemediations(incidentId?: string): Promise<Remediation[]> {
  const { data } = await api.get("/remediations", { params: incidentId ? { incident_id: incidentId } : {} });
  return data;
}

export async function approveRemediation(id: string): Promise<Remediation> {
  const { data } = await api.post(`/remediations/${id}/approve`);
  return data;
}

export async function executeRemediation(id: string): Promise<Remediation> {
  const { data } = await api.post(`/remediations/${id}/execute`);
  return data;
}
