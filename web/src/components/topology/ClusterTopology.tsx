import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  Node,
  Edge,
} from "reactflow";
import "reactflow/dist/style.css";

const initialNodes: Node[] = [
  {
    id: "frontend",
    position: { x: 250, y: 0 },
    data: { label: "Frontend" },
    style: { background: "hsl(160 84% 39%)", color: "#fff", border: "none", borderRadius: "8px", padding: "12px 24px" },
  },
  {
    id: "api",
    position: { x: 250, y: 120 },
    data: { label: "API Gateway" },
    style: { background: "hsl(217 91% 60%)", color: "#fff", border: "none", borderRadius: "8px", padding: "12px 24px" },
  },
  {
    id: "checkout",
    position: { x: 100, y: 240 },
    data: { label: "Checkout" },
    style: { background: "hsl(160 84% 39%)", color: "#fff", border: "none", borderRadius: "8px", padding: "12px 24px" },
  },
  {
    id: "payment",
    position: { x: 400, y: 240 },
    data: { label: "Payment" },
    style: { background: "hsl(0 72% 51%)", color: "#fff", border: "none", borderRadius: "8px", padding: "12px 24px" },
  },
  {
    id: "database",
    position: { x: 250, y: 360 },
    data: { label: "Database" },
    style: { background: "hsl(160 84% 39%)", color: "#fff", border: "none", borderRadius: "8px", padding: "12px 24px" },
  },
];

const initialEdges: Edge[] = [
  { id: "e-fe-api", source: "frontend", target: "api", animated: true, style: { stroke: "hsl(215 28% 40%)" } },
  { id: "e-api-checkout", source: "api", target: "checkout", animated: true, style: { stroke: "hsl(215 28% 40%)" } },
  { id: "e-api-payment", source: "api", target: "payment", animated: true, style: { stroke: "hsl(215 28% 40%)" } },
  { id: "e-checkout-db", source: "checkout", target: "database", animated: true, style: { stroke: "hsl(215 28% 40%)" } },
  { id: "e-payment-db", source: "payment", target: "database", animated: true, style: { stroke: "hsl(215 28% 40%)" } },
];

export function ClusterTopology() {
  const navigate = useNavigate();

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      navigate(`/incidents?service=${node.id}`);
    },
    [navigate]
  );

  return (
    <div className="h-[600px] rounded-lg border border-border bg-card">
      <ReactFlow
        nodes={initialNodes}
        edges={initialEdges}
        onNodeClick={onNodeClick}
        fitView
        attributionPosition="bottom-right"
      >
        <Background color="hsl(215 28% 17%)" gap={20} />
        <Controls className="bg-card border-border" />
        <MiniMap
          style={{ backgroundColor: "hsl(215 28% 8%)" }}
          maskColor="hsl(215 28% 8% / 0.7)"
        />
      </ReactFlow>
    </div>
  );
}
