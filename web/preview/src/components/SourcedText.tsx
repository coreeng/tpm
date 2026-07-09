import type { ReactNode } from "react";
import type { Sourced } from "../types";
import { SourceMarker } from "./SourceMarker";

interface SourcedInlineProps {
  value: Sourced<string | number>;
  className?: string;
}

export function SourcedInline({ value, className = "" }: SourcedInlineProps) {
  if (value.value === "") {
    return null;
  }
  return (
    <span className={`inline-flex items-center gap-1.5 ${className}`}>
      <span>{value.value}</span>
      <SourceMarker source={value.source} />
    </span>
  );
}

interface HeadingWithSourceProps {
  children: ReactNode;
  source?: Sourced<unknown>["source"];
  level?: "h1" | "h2" | "h3" | "h4";
  className?: string;
}

export function HeadingWithSource({
  children,
  source,
  level = "h2",
  className = "",
}: HeadingWithSourceProps) {
  const Tag = level;
  return (
    <div className={`flex min-w-0 items-start gap-2 ${className}`}>
      <Tag className="min-w-0">{children}</Tag>
      <SourceMarker source={source} />
    </div>
  );
}

interface ProseBlockProps {
  text: Sourced<string>;
  label?: string;
}

export function ProseBlock({ text, label = "Description" }: ProseBlockProps) {
  if (!text.value) {
    return null;
  }
  return (
    <div className="space-y-1">
      <div className="flex items-center gap-1.5 text-[11px] font-bold uppercase tracking-wider text-slate-500">
        <span>{label}</span>
        <SourceMarker source={text.source} />
      </div>
      <pre className="text-sm leading-6 text-slate-700">{text.value}</pre>
    </div>
  );
}
