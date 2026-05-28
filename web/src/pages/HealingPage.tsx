import { HealingCenter } from "@/components/healing/HealingCenter";

export default function HealingPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl italic text-stone-900">Self-Healing</h1>
        <p className="mt-1 text-sm text-stone-500">Automated remediation audit trail.</p>
      </div>
      <HealingCenter />
    </div>
  );
}
