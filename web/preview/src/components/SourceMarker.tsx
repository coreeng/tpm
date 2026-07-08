import { FileText } from "lucide-react";
import { useCallback, useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import type { SourceRef } from "../types";

interface SourceMarkerProps {
  source?: SourceRef;
}

interface TooltipPosition {
  top: number;
  left: number;
  arrowLeft: number;
  placement: "top" | "bottom";
}

export function SourceMarker({ source }: SourceMarkerProps) {
  const buttonRef = useRef<HTMLButtonElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);
  const tooltipId = useId();
  const [open, setOpen] = useState(false);
  const [position, setPosition] = useState<TooltipPosition | null>(null);

  const updatePosition = useCallback(() => {
    const button = buttonRef.current;
    if (!button) {
      return;
    }

    const rect = button.getBoundingClientRect();
    const viewportPadding = 16;
    const fallbackWidth = Math.min(448, window.innerWidth - viewportPadding * 2);
    const tooltipWidth = tooltipRef.current?.offsetWidth || fallbackWidth;
    const tooltipHeight = tooltipRef.current?.offsetHeight || 132;
    const triggerCenter = rect.left + rect.width / 2;
    const maxLeft = Math.max(viewportPadding, window.innerWidth - tooltipWidth - viewportPadding);
    const left = Math.min(Math.max(triggerCenter - tooltipWidth / 2, viewportPadding), maxLeft);

    let top = rect.bottom + 10;
    let placement: TooltipPosition["placement"] = "bottom";
    if (top + tooltipHeight > window.innerHeight - viewportPadding && rect.top - tooltipHeight - 10 >= viewportPadding) {
      top = rect.top - tooltipHeight - 10;
      placement = "top";
    }

    setPosition({
      top,
      left,
      arrowLeft: triggerCenter - left,
      placement,
    });
  }, []);

  useLayoutEffect(() => {
    if (!open) {
      return;
    }
    updatePosition();
  }, [open, source?.file, source?.property, updatePosition]);

  useEffect(() => {
    if (!open) {
      return;
    }

    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, true);
    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition, true);
    };
  }, [open, updatePosition]);

  if (!source?.file) {
    return null;
  }

  const label = source.property
    ? `Source file ${source.file}, YAML property ${source.property}`
    : `Source file ${source.file}`;

  return (
    <span
      className="relative inline-flex shrink-0 align-middle"
      onBlur={() => setOpen(false)}
      onFocus={() => setOpen(true)}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
    >
      <button
        ref={buttonRef}
        type="button"
        aria-label={label}
        aria-describedby={open ? tooltipId : undefined}
        className="inline-flex size-5 items-center justify-center rounded-full border border-slate-300 bg-white text-slate-500 shadow-sm transition hover:border-sky-500 hover:bg-sky-50 hover:text-sky-700 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:ring-offset-2"
        onKeyDown={(event) => {
          if (event.key === "Escape") {
            setOpen(false);
          }
        }}
      >
        <FileText aria-hidden="true" className="size-3.5" strokeWidth={2} />
      </button>
      {open && typeof document !== "undefined"
        ? createPortal(
            <div
              ref={tooltipRef}
              id={tooltipId}
              role="tooltip"
              className="pointer-events-none fixed z-[1000] w-[28rem] max-w-[calc(100vw-2rem)] rounded-lg border border-slate-200 bg-white p-3 text-left text-xs text-slate-900 shadow-xl"
              style={{
                left: position?.left ?? 0,
                opacity: position ? 1 : 0,
                top: position?.top ?? 0,
              }}
            >
              <span
                className={
                  position?.placement === "top"
                    ? "absolute -bottom-1.5 size-3 -translate-x-1/2 rotate-45 border-b border-r border-slate-200 bg-white"
                    : "absolute -top-1.5 size-3 -translate-x-1/2 rotate-45 border-l border-t border-slate-200 bg-white"
                }
                style={{ left: position?.arrowLeft ?? 0 }}
              />
              <span className="block text-[10px] font-bold uppercase tracking-wider text-slate-500">
                File
              </span>
              <code className="mt-1 block break-words font-mono text-[12px] leading-5 text-slate-900">
                {source.file}
              </code>
              {source.property ? (
                <>
                  <span className="mt-2 block text-[10px] font-bold uppercase tracking-wider text-slate-500">
                    YAML property
                  </span>
                  <code className="mt-1 block break-words font-mono text-[12px] leading-5 text-slate-900">
                    {source.property}
                  </code>
                </>
              ) : null}
            </div>,
            document.body,
          )
        : null}
    </span>
  );
}
