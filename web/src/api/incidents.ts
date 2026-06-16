import api from "./client";
import type { Incident, TimelineEntry } from "@/types";

export async function fetchIncidents(params?: Record<string, string>): Promise<Incident[]> {
  const { data } = await api.get("/incidents", { params });
  return data;
}

export async function fetchIncident(id: string): Promise<Incident> {
  const { data } = await api.get(`/incidents/${id}`);
  return data;
}

export async function createIncident(incident: Partial<Incident>): Promise<Incident> {
  const { data } = await api.post("/incidents", incident);
  return data;
}

export async function acknowledgeIncident(id: string): Promise<Incident> {
  const { data } = await api.put(`/incidents/${id}/acknowledge`);
  return data;
}

export async function resolveIncident(id: string): Promise<Incident> {
  const { data } = await api.put(`/incidents/${id}/resolve`);
  return data;
}

export async function fetchTimeline(id: string): Promise<TimelineEntry[]> {
  const { data } = await api.get(`/incidents/${id}/timeline`);
  return data;
}
