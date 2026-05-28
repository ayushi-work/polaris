import { ChaosLab } from "@/components/chaos/ChaosLab";

export default function ChaosLabPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl italic text-stone-900">Chaos Lab</h1>
        <p className="mt-1 text-sm text-stone-500">Inject controlled failures to test resilience.</p>
      </div>
      <ChaosLab />
    </div>
  );
}
