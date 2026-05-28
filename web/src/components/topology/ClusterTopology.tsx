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
    style: {
      background: "#fff",
      color: "#1c1917",
      border: "1.5px solid #d6d3d1",
      borderRadius: "10px",
      padding: "10px 22px",
      fontSize: "13px",
      fontWeight: 500,
    },
  },
  {
    id: "api",
    position: { x: 250, y: 120 },
    data: { label: "API Gateway" },
    style: {
      background: "#fff",
      color: "#1c1917",
      border: "1.5px solid #d6d3d1",
      borderRadius: "10px",
      padding: "10px 22px",
      fontSize: "13px",
      fontWeight: 500,
    },
  },
  {
    id: "checkout",
    position: { x: 100, y: 240 },
    data: { label: "Checkout" },
    style: {
      background: "#fff",
      color: "#1c1917",
      border: "1.5px solid #d6d3d1",
      borderRadius: "10px",
      padding: "10px 22px",
      fontSize: "13px",
      fontWeight: 500,
    },
  },
  {
    id: "payment",
    position: { x: 400, y: 240 },
    data: { label: "Payment" },
    style: {
      background: "#fef2f2",
      color: "#991b1b",
      border: "1.5px solid #fecaca",
      borderRadius: "10px",
      padding: "10px 22px",
      fontSize: "13px",
      fontWeight: 500,
    },
  },
  {
    id: "database",
    position: { x: 250, y: 360 },
    data: { label: "Database" },
    style: {
      background: "#fff",
      color: "#1c1917",
      border: "1.5px solid #d6d3d1",
      borderRadius: "10px",
      padding: "10px 22px",
      fontSize: "13px",
      fontWeight: 500,
    },
  },
];

const initialEdges: Edge[] = [
  { id: "e-fe-api", source: "frontend", target: "api", animated: true, style: { stroke: "#d6d3d1", strokeWidth: 2 } },
  { id: "e-api-checkout", source: "api", target: "checkout", animated: true, style: { stroke: "#d6d3d1", strokeWidth: 2 } },
  { id: "e-api-payment", source: "api", target: "payment", animated: true, style: { stroke: "#d6d3d1", strokeWidth: 2 } },
  { id: "e-checkout-db", source: "checkout", target: "database", animated: true, style: { stroke: "#d6d3d1", strokeWidth: 2 } },
  { id: "e-payment-db", source: "payment", target: "database", animated: true, style: { stroke: "#d6d3d1", strokeWidth: 2 } },
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
    <div className="h-full w-full">
      <ReactFlow
        nodes={initialNodes}
        edges={initialEdges}
        onNodeClick={onNodeClick}
        fitView
        attributionPosition="bottom-right"
      >
        <Background color="#e7e5e4" gap={20} />
        <Controls className="!bg-white !border-stone-200 !rounded-lg !shadow-sm" />
        <MiniMap
          style={{ backgroundColor: "#fafaf9" }}
          maskColor="rgba(0,0,0,0.04)"
          nodeColor="#d6d3d1"
        />
      </ReactFlow>
    </div>
  );
}
