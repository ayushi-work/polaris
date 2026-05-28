import { PostmortemGenerator } from "@/components/postmortems/PostmortemGenerator";

export default function PostmortemsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl italic text-stone-900">Postmortems</h1>
        <p className="mt-1 text-sm text-stone-500">Generate downloadable incident reports.</p>
      </div>
      <PostmortemGenerator />
    </div>
  );
}
