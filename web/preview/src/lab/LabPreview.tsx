import {
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  ChevronDown,
  Clock,
  FlaskConical,
  KeyRound,
  Server,
  Target,
  XCircle,
} from "lucide-react";
import { useState } from "react";
import { HeadingWithSource, ProseBlock, SourcedInline } from "../components/SourcedText";
import { SourceMarker } from "../components/SourceMarker";
import type { LabChallenge, LabPreview as LabPreviewPayload, LabProgress } from "../types";

interface LabPreviewProps {
  data: LabPreviewPayload;
}

export function LabPreview({ data }: LabPreviewProps) {
  const [activeChallengeKey, setActiveChallengeKey] = useState<string | null>(() =>
    data.challenges[0] ? labChallengeKey(data.challenges[0], 0) : null,
  );
  const selectedIndex = activeChallengeKey
    ? data.challenges.findIndex((challenge, index) => labChallengeKey(challenge, index) === activeChallengeKey)
    : -1;
  const activeIndex = selectedIndex >= 0 ? selectedIndex : 0;
  const activeChallenge = data.challenges[activeIndex] ?? data.challenges[0];
  const labPalette = progressHeaderPalette(data.progress);
  const completedChallenges = data.challenges.filter((challenge) => challenge.progress.state === "complete").length;
  const totalGoals = data.challenges.reduce((total, challenge) => total + challenge.goals.length, 0);
  const completedGoals = data.challenges.reduce(
    (total, challenge) => total + challenge.goals.filter((goal) => goal.progress.state === "complete").length,
    0,
  );

  return (
    <main className="min-h-screen bg-slate-50">
      <section className={`border-b ${labPalette.border} ${labPalette.surface}`}>
        <div className="mx-auto flex max-w-7xl flex-col gap-5 px-6 py-7 lg:px-8">
          <div className="flex flex-wrap items-start justify-between gap-5">
            <div className="max-w-3xl space-y-3">
              <div className={`flex items-center gap-2 text-xs font-bold uppercase tracking-wider ${labPalette.eyebrow}`}>
                <FlaskConical className="size-4" />
                <span>Lab preview</span>
              </div>
              <HeadingWithSource
                className="[&_h1]:text-4xl [&_h1]:font-semibold [&_h1]:tracking-tight [&_h1]:text-white"
                level="h1"
                source={data.title.source}
              >
                {data.title.value}
              </HeadingWithSource>
              <div className="flex flex-wrap items-center gap-2 text-sm">
                <SourcedInline className={labPalette.metaPill} value={data.code} />
                {data.timeLimit.value ? (
                  <span className={`inline-flex items-center gap-1.5 ${labPalette.metaPill}`}>
                    <Clock className="size-4" />
                    <SourcedInline value={data.timeLimit} />
                  </span>
                ) : null}
              </div>
              {data.description.value ? (
                <p className={`max-w-4xl text-base leading-7 ${labPalette.description}`}>
                  {data.description.value} <SourceMarker source={data.description.source} />
                </p>
              ) : null}
            </div>
            <LabCompletionBadge
              completedChallenges={completedChallenges}
              completedGoals={completedGoals}
              totalChallenges={data.challenges.length}
              totalGoals={totalGoals}
            />
          </div>
          {data.statusError ? (
            <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
              <div className="flex items-start gap-2">
                <AlertCircle className="mt-0.5 size-4 shrink-0" />
                <div>
                  <div className="font-semibold">Status unavailable</div>
                  <div className="mt-1">{data.statusError}</div>
                </div>
              </div>
            </div>
          ) : null}
        </div>
      </section>

      {data.runtime ? (
        <div className="mx-auto max-w-7xl px-6 pt-6 lg:px-8">
          <RuntimePanel runtime={data.runtime} />
        </div>
      ) : null}

      <div className="mx-auto grid max-w-7xl gap-6 px-6 py-6 lg:grid-cols-[320px_1fr] lg:px-8">
        <aside className="lg:sticky lg:top-6 lg:self-start">
          <nav className="rounded-lg border border-slate-200 bg-white p-3 shadow-sm">
            <div className="mb-3 flex items-center gap-2 px-2 text-sm font-semibold text-slate-900">
              <Target className="size-4 text-emerald-700" />
              <span>Challenges</span>
            </div>
            <div className="space-y-2">
              {data.challenges.map((challenge, index) => (
                <button
                  className={`grid w-full grid-cols-[2rem_1fr] gap-2 rounded-lg border px-3 py-2 text-left transition ${progressListPalette(
                    challenge.progress,
                    index === activeIndex,
                  )}`}
                  key={labChallengeKey(challenge, index)}
                  onClick={() => setActiveChallengeKey(labChallengeKey(challenge, index))}
                  type="button"
                >
                  <span className={progressListIndexPalette(challenge.progress, index === activeIndex)}>
                    {index + 1}.
                  </span>
                  <span>
                    <span className="block text-sm font-semibold text-slate-900">{challenge.title.value}</span>
                    <span className="mt-1 flex flex-wrap items-center gap-2">
                      <span className="text-xs text-slate-500">{challenge.goals.length} goals</span>
                      <ProgressBadge compact progress={challenge.progress} />
                    </span>
                  </span>
                </button>
              ))}
            </div>
          </nav>
        </aside>
        <section>{activeChallenge ? <ChallengeView challenge={activeChallenge} index={activeIndex} /> : <EmptyState />}</section>
      </div>
    </main>
  );
}

function LabCompletionBadge({
  completedChallenges,
  completedGoals,
  totalChallenges,
  totalGoals,
}: {
  completedChallenges: number;
  completedGoals: number;
  totalChallenges: number;
  totalGoals: number;
}) {
  return (
    <div className="flex flex-wrap items-center gap-2 rounded-full bg-white px-4 py-2 text-xs font-bold uppercase tracking-wide text-slate-950 shadow-sm ring-1 ring-white/70">
      <span>
        {completedChallenges}/{totalChallenges} challenges completed
      </span>
      <span className="h-4 w-px bg-slate-300" />
      <span>
        {completedGoals}/{totalGoals} goals completed
      </span>
    </div>
  );
}

function RuntimePanel({ runtime }: { runtime: NonNullable<LabPreviewPayload["runtime"]> }) {
  const [revealToken, setRevealToken] = useState(false);
  return (
    <details className="overflow-hidden rounded-lg border border-cyan-200 bg-white shadow-sm">
      <summary className="flex cursor-pointer list-none items-center justify-between gap-4 bg-cyan-700 px-5 py-4 text-base font-semibold text-white sm:text-sm">
        <span className="inline-flex items-center gap-2">
          <Server className="size-5 shrink-0 sm:size-4" />
          Runtime details
        </span>
        <ChevronDown className="size-5 shrink-0 text-cyan-100 sm:size-4" />
      </summary>
      <div className="grid gap-3 bg-cyan-50/40 p-5 text-sm sm:grid-cols-2">
        <RuntimeValue label="Run" value={runtime.runId} />
        <RuntimeValue label="System namespace" value={runtime.systemNamespace} />
        <RuntimeValue label="Workspace namespace" value={runtime.workspaceNamespace} />
        <RuntimeValue label="Registry" value={runtime.registryUrl} />
        <RuntimeValue label="Registry username" value={runtime.registryUsername} />
        <div className="rounded-md border border-slate-200 bg-white p-3">
          <div className="mb-1 flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-slate-500">
            <KeyRound className="size-3.5" />
            <span>Registry token</span>
          </div>
          <div className="flex items-center justify-between gap-3">
            <code className="min-w-0 truncate font-mono text-xs text-slate-800">
              {revealToken ? runtime.registryToken : "••••••••••••••••"}
            </code>
            <button
              className="rounded border border-slate-300 px-2 py-1 text-xs font-medium text-slate-700 hover:bg-slate-50"
              onClick={() => setRevealToken((current) => !current)}
              type="button"
            >
              {revealToken ? "Hide" : "Reveal"}
            </button>
          </div>
        </div>
      </div>
    </details>
  );
}

function RuntimeValue({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-slate-200 bg-white p-3">
      <div className="mb-1 text-xs font-bold uppercase tracking-wider text-slate-500">{label}</div>
      <code className="block truncate font-mono text-xs text-slate-800">{value}</code>
    </div>
  );
}

function ChallengeView({ challenge, index }: { challenge: LabChallenge; index: number }) {
  const challengePalette = progressHeaderPalette(challenge.progress);
  const completedGoals = challenge.goals.filter((goal) => goal.progress.state === "complete").length;

  return (
    <article className={`overflow-hidden rounded-lg border bg-white shadow-sm ${challengePalette.cardBorder}`}>
      <div className={`${challengePalette.surface} px-6 py-5 text-white`}>
        <div className="mb-2 flex flex-wrap items-center justify-between gap-3">
          <div className={`flex items-center gap-2 text-sm font-semibold sm:text-xs ${challengePalette.eyebrow}`}>
            <Target className="size-4 shrink-0" />
            <span>Challenge {index + 1}</span>
          </div>
          <GoalCountBadge completed={completedGoals} total={challenge.goals.length} />
        </div>
        <HeadingWithSource
          className="mt-1 [&_h2]:text-3xl [&_h2]:font-semibold [&_h2]:tracking-tight [&_h2]:text-white"
          source={challenge.title.source}
        >
          {challenge.title.value}
        </HeadingWithSource>
      </div>
      <div className="space-y-5 p-6">
        <ProseBlock text={challenge.description} />
        <ProgressDetail progress={challenge.progress} />
        <div className="space-y-3">
          {challenge.goals.map((goal, goalIndex) => (
            <GoalCard goal={goal} index={goalIndex} key={labGoalKey(goal, goalIndex)} />
          ))}
        </div>
        <ProseBlock label="Success message" text={challenge.successMessage} />
      </div>
    </article>
  );
}

function GoalCard({ goal, index }: { goal: LabChallenge["goals"][number]; index: number }) {
  const goalPalette = progressHeaderPalette(goal.progress);

  return (
    <div className={`overflow-hidden rounded-md border bg-white shadow-sm ${goalPalette.cardBorder}`}>
      <div className={`${goalPalette.surface} px-4 py-3 text-white`}>
        <div className="mb-2 flex flex-wrap items-center justify-between gap-3">
          <div className={`flex items-center gap-2 text-sm font-semibold sm:text-xs ${goalPalette.eyebrow}`}>
            <CheckCircle2 className="size-4 shrink-0" />
            <span>Goal {index + 1}</span>
          </div>
          <ProgressBadge onDark progress={goal.progress} />
        </div>
        <HeadingWithSource
          className="[&_h4]:text-xl [&_h4]:font-semibold [&_h4]:text-white"
          level="h4"
          source={goal.title.source}
        >
          {goal.title.value}
        </HeadingWithSource>
      </div>
      {goal.description.value || goal.progress.message ? (
        <div className="space-y-3 p-4">
          <ProgressDetail progress={goal.progress} />
          {goal.description.value ? (
            <p className="text-base text-slate-600 sm:text-sm">
              {goal.description.value} <SourceMarker source={goal.description.source} />
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function GoalCountBadge({ completed, total }: { completed: number; total: number }) {
  return (
    <span className="rounded-full bg-white px-3 py-1 text-xs font-bold uppercase tracking-wide text-slate-950 shadow-sm ring-1 ring-white/70">
      {completed}/{total} goals completed
    </span>
  );
}

function ProgressBadge({
  compact = false,
  onDark = false,
  progress,
}: {
  compact?: boolean;
  onDark?: boolean;
  progress: LabProgress;
}) {
  const Icon = progress.state === "complete" ? CheckCircle2 : progress.state === "incomplete" ? XCircle : AlertTriangle;
  const palette = progressPalette(progress, onDark);
  const iconPalette = progressIconPalette(progress);
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full font-bold uppercase tracking-wide ring-1 ${
        compact ? "px-2 py-0.5 text-[11px]" : "px-3 py-1 text-xs"
      } ${palette}`}
    >
      <Icon className={`${compact ? "size-3.5" : "size-4"} shrink-0 ${iconPalette}`} />
      <span>{progressBadgeLabel(progress)}</span>
    </span>
  );
}

function ProgressDetail({ progress }: { progress: LabProgress }) {
  if (!progress.reason && !progress.message) {
    return null;
  }
  return (
    <div className={`rounded-md border px-3 py-2 text-sm ${progressDetailPalette(progress)}`}>
      <div className="font-semibold">{progress.reason || progress.label}</div>
      {progress.message ? <div className="mt-1">{progress.message}</div> : null}
      <code className="mt-2 block truncate font-mono text-[12px] opacity-75">{progress.conditionType}</code>
    </div>
  );
}

function progressPalette(progress: LabProgress, onDark: boolean) {
  if (onDark) {
    switch (progress.state) {
      case "complete":
        return "bg-white text-slate-950 shadow-sm ring-white/70";
      case "incomplete":
        return "bg-white text-slate-950 shadow-sm ring-white/70";
      default:
        return "bg-white text-slate-950 shadow-sm ring-white/70";
    }
  }
  switch (progress.state) {
    case "complete":
      return "bg-emerald-50 text-emerald-900 ring-emerald-200";
    case "incomplete":
      return "bg-rose-50 text-rose-900 ring-rose-200";
    default:
      return "bg-slate-100 text-slate-700 ring-slate-200";
  }
}

function progressBadgeLabel(progress: LabProgress) {
  switch (progress.state) {
    case "complete":
      return "COMPLETE";
    case "incomplete":
      return "INCOMPLETE";
    default:
      return "NOT REPORTED";
  }
}

function progressIconPalette(progress: LabProgress) {
  switch (progress.state) {
    case "complete":
      return "text-emerald-600";
    case "incomplete":
      return "text-rose-600";
    default:
      return "text-slate-500";
  }
}

function progressHeaderPalette(progress: LabProgress) {
  switch (progress.state) {
    case "complete":
      return {
        border: "border-emerald-800",
        cardBorder: "border-emerald-300",
        description: "text-emerald-50",
        eyebrow: "text-emerald-100",
        metaPill: "rounded-full bg-white/15 px-3 py-1 font-medium text-white ring-1 ring-white/25",
        surface: "bg-emerald-700",
      };
    case "incomplete":
      return {
        border: "border-rose-800",
        cardBorder: "border-rose-300",
        description: "text-rose-50",
        eyebrow: "text-rose-100",
        metaPill: "rounded-full bg-white/15 px-3 py-1 font-medium text-white ring-1 ring-white/25",
        surface: "bg-rose-700",
      };
    default:
      return {
        border: "border-slate-800",
        cardBorder: "border-slate-300",
        description: "text-slate-100",
        eyebrow: "text-slate-200",
        metaPill: "rounded-full bg-white/15 px-3 py-1 font-medium text-white ring-1 ring-white/25",
        surface: "bg-slate-700",
      };
  }
}

function progressListPalette(progress: LabProgress, active: boolean) {
  if (active) {
    switch (progress.state) {
      case "complete":
        return "border-emerald-300 bg-emerald-50";
      case "incomplete":
        return "border-rose-300 bg-rose-50";
      default:
        return "border-slate-300 bg-slate-50";
    }
  }
  return "border-transparent hover:bg-slate-50";
}

function progressListIndexPalette(progress: LabProgress, active: boolean) {
  if (!active) {
    return "font-semibold text-slate-500";
  }
  switch (progress.state) {
    case "complete":
      return "font-semibold text-emerald-700";
    case "incomplete":
      return "font-semibold text-rose-700";
    default:
      return "font-semibold text-slate-700";
  }
}

function progressDetailPalette(progress: LabProgress) {
  switch (progress.state) {
    case "complete":
      return "border-emerald-200 bg-emerald-50 text-emerald-900";
    case "incomplete":
      return "border-rose-200 bg-rose-50 text-rose-900";
    default:
      return "border-slate-200 bg-slate-50 text-slate-700";
  }
}

function labChallengeKey(challenge: LabChallenge, index: number) {
  return challenge.code.value || `challenge-${index}`;
}

function labGoalKey(goal: LabChallenge["goals"][number], index: number) {
  return goal.code.value || `goal-${index}`;
}

function EmptyState() {
  return (
    <div className="rounded-lg border border-dashed border-slate-300 bg-white p-10 text-center text-slate-600">
      No challenges found in this lab.
    </div>
  );
}
