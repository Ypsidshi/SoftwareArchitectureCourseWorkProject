export default function ErrorAlert({ message, className = "" }: { message: string; className?: string }) {
  if (!message) return null;
  return (
    <div className={`rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 ${className}`.trim()}>
      {message}
    </div>
  );
}
