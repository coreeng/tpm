import {
  BadgeCheck,
  BookOpen,
  CheckCircle2,
  Circle,
  ClipboardList,
  Clock,
  FlaskConical,
  HelpCircle,
  Layers,
  ListChecks,
  PlayCircle,
} from "lucide-react";
import type { ReactNode } from "react";
import { useMemo, useState } from "react";
import { HeadingWithSource, ProseBlock, SourcedInline } from "../components/SourcedText";
import { SourceMarker } from "../components/SourceMarker";
import type {
  ModuleChapter,
  ModuleChallenge,
  ModuleGoal,
  ModuleLab,
  ModulePreview as ModulePreviewPayload,
  ModuleQuestion,
  ModuleQuiz,
  ModuleSection,
} from "../types";

interface ModulePreviewProps {
  data: ModulePreviewPayload;
}

export function ModulePreview({ data }: ModulePreviewProps) {
  const [activeChapterKey, setActiveChapterKey] = useState<string | null>(() =>
    data.chapters[0] ? moduleChapterKey(data.chapters[0], 0) : null,
  );
  const selectedIndex = activeChapterKey
    ? data.chapters.findIndex((chapter, index) => moduleChapterKey(chapter, index) === activeChapterKey)
    : -1;
  const activeIndex = selectedIndex >= 0 ? selectedIndex : 0;
  const activeChapter = data.chapters[activeIndex] ?? data.chapters[0];

  return (
    <main className="min-h-screen bg-slate-50">
      <section className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-7xl flex-col gap-5 px-6 py-7 lg:px-8">
          <div className="flex flex-wrap items-start justify-between gap-5">
            <div className="max-w-3xl space-y-3">
              <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-sky-700">
                <BookOpen className="size-4" />
                <span>Module preview</span>
              </div>
              <HeadingWithSource
                className="[&_h1]:text-4xl [&_h1]:font-semibold [&_h1]:tracking-tight [&_h1]:text-slate-950"
                level="h1"
                source={data.title.source}
              >
                {data.title.value}
              </HeadingWithSource>
              <div className="flex flex-wrap items-center gap-2 text-sm text-slate-600">
                <SourcedInline
                  className="rounded-full bg-sky-50 px-3 py-1 font-semibold uppercase tracking-wide text-sky-800"
                  value={data.level}
                />
                <SourcedInline
                  className="rounded-full bg-slate-100 px-3 py-1 font-medium text-slate-700"
                  value={data.code}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
              <Stat label="Chapters" value={data.stats.chapters} icon={<BookOpen className="size-4" />} />
              <Stat label="Sections" value={data.stats.sections} icon={<Layers className="size-4" />} />
              <Stat label="Quizzes" value={data.stats.quizzes} icon={<ClipboardList className="size-4" />} />
              <Stat label="Labs" value={data.stats.labs} icon={<FlaskConical className="size-4" />} />
            </div>
          </div>
          <ProseBlock text={data.description} />
        </div>
      </section>

      <div className="mx-auto grid max-w-7xl gap-6 px-6 py-6 lg:grid-cols-[320px_1fr] lg:px-8">
        <aside className="lg:sticky lg:top-6 lg:self-start">
          <nav className="rounded-lg border border-slate-200 bg-white p-3 shadow-sm">
            <div className="mb-3 flex items-center gap-2 px-2 text-sm font-semibold text-slate-900">
              <ListChecks className="size-4 text-sky-700" />
              <span>Outline</span>
            </div>
            <div className="space-y-2">
              {data.chapters.map((chapter, index) => (
                <ChapterNavItem
                  active={index === activeIndex}
                  chapter={chapter}
                  index={index}
                  key={moduleChapterKey(chapter, index)}
                  onSelect={() => setActiveChapterKey(moduleChapterKey(chapter, index))}
                />
              ))}
            </div>
          </nav>
        </aside>

        <section>{activeChapter ? <ChapterView chapter={activeChapter} index={activeIndex} /> : <EmptyState />}</section>
      </div>
    </main>
  );
}

function Stat({ label, value, icon }: { label: string; value: number; icon: ReactNode }) {
  return (
    <div className="min-w-24 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
      <div className="flex items-center gap-1.5 text-xs font-medium text-slate-500">
        {icon}
        <span>{label}</span>
      </div>
      <div className="mt-1 text-2xl font-semibold tabular-nums text-slate-950">{value}</div>
    </div>
  );
}

function ChapterNavItem({
  active,
  chapter,
  index,
  onSelect,
}: {
  active: boolean;
  chapter: ModuleChapter;
  index: number;
  onSelect: () => void;
}) {
  return (
    <div
      className={
        active
          ? "rounded-lg border border-sky-200 bg-sky-50"
          : "rounded-lg border border-transparent bg-white"
      }
    >
      <button
        className="grid w-full grid-cols-[2rem_1fr] gap-2 rounded-lg px-3 py-2 text-left transition hover:bg-slate-50"
        onClick={onSelect}
        type="button"
      >
        <span className={active ? "font-semibold text-sky-700" : "font-semibold text-slate-500"}>
          {index + 1}.
        </span>
        <span>
          <span className="block text-sm font-semibold text-slate-900">{chapter.title.value}</span>
          <span className="block text-xs text-slate-500">
            {chapter.sections.length} sections
            {chapter.multipleChoiceAssessments.length ? ` · ${chapter.multipleChoiceAssessments.length} quiz` : ""}
            {chapter.labs.length ? ` · ${chapter.labs.length} lab` : ""}
          </span>
        </span>
      </button>
      {active && chapter.sections.length ? (
        <div className="space-y-1 border-t border-sky-100 px-11 py-2">
          {chapter.sections.map((section, sectionIndex) => (
            <a
              className="block rounded px-2 py-1 text-xs leading-5 text-slate-600 hover:bg-white hover:text-sky-700"
              href={`#section-${index}-${sectionIndex}`}
              key={moduleSectionKey(section, sectionIndex)}
            >
              {sectionIndex + 1}. {section.title.value}
            </a>
          ))}
        </div>
      ) : null}
    </div>
  );
}

function ChapterView({ chapter, index }: { chapter: ModuleChapter; index: number }) {
  const contentCount = useMemo(
    () => chapter.sections.length + chapter.multipleChoiceAssessments.length + chapter.labs.length,
    [chapter],
  );

  return (
    <article className="space-y-5">
      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="bg-sky-700 px-6 py-5 text-white">
          <div className="text-sm font-semibold text-sky-100 sm:text-xs">Chapter {index + 1}</div>
          <HeadingWithSource
            className="mt-1 [&_h2]:text-3xl [&_h2]:font-semibold [&_h2]:tracking-tight [&_h2]:text-white"
            source={chapter.title.source}
          >
            {chapter.title.value}
          </HeadingWithSource>
          <div className="mt-3 flex flex-wrap gap-2 text-xs text-white">
            <span className="rounded-full bg-white/15 px-2.5 py-1 font-medium ring-1 ring-white/20">{contentCount} items</span>
            {chapter.isDraft ? (
              <span className="rounded-full bg-amber-300 px-2.5 py-1 font-semibold text-amber-950">Draft</span>
            ) : null}
          </div>
        </div>
        <div className="space-y-4 p-6">
          <ProseBlock text={chapter.description} />
          {chapter.bannerVideo.value ? (
            <a className="inline-flex items-center gap-2 text-sm font-medium text-sky-700" href={chapter.bannerVideo.value}>
              <PlayCircle className="size-4" />
              <span>Video link</span>
              <SourceMarker source={chapter.bannerVideo.source} />
            </a>
          ) : null}
        </div>
      </section>

      <div className="space-y-4">
        {chapter.sections.map((section, sectionIndex) => (
          <SectionBlock
            chapterIndex={index}
            index={sectionIndex}
            key={moduleSectionKey(section, sectionIndex)}
            section={section}
          />
        ))}
        {chapter.labs.map((lab, labIndex) => (
          <LabBlock key={moduleLabKey(lab, labIndex)} lab={lab} />
        ))}
        {chapter.multipleChoiceAssessments.map((quiz, quizIndex) => (
          <QuizBlock key={moduleQuizKey(quiz, quizIndex)} quiz={quiz} />
        ))}
      </div>
    </article>
  );
}

function moduleChapterKey(chapter: ModuleChapter, index: number) {
  return chapter.code.value || `chapter-${chapter.index || index}`;
}

function moduleSectionKey(section: ModuleSection, index: number) {
  return section.code.value || `section-${section.index || index}`;
}

function moduleLabKey(lab: ModuleLab, index: number) {
  return lab.code.value || `lab-${lab.index || index}`;
}

function moduleQuizKey(quiz: ModuleQuiz, index: number) {
  return quiz.code.value || `quiz-${quiz.index || index}`;
}

function moduleChallengeKey(challenge: ModuleChallenge, index: number) {
  return challenge.code.value || `challenge-${challenge.index || index}`;
}

function moduleGoalKey(goal: ModuleGoal, index: number) {
  return goal.code.value || `goal-${goal.index || index}`;
}

function moduleQuestionKey(question: ModuleQuestion, index: number) {
  return question.code.value || `question-${question.index || index}`;
}

function SectionBlock({
  chapterIndex,
  index,
  section,
}: {
  chapterIndex: number;
  index: number;
  section: ModuleSection;
}) {
  return (
    <section
      className="overflow-hidden rounded-lg border border-teal-200 bg-white shadow-sm"
      id={`section-${chapterIndex}-${index}`}
    >
      <div className="bg-teal-700 px-5 py-4 text-white">
        <div className="mb-1 flex items-center gap-2 text-sm font-semibold text-teal-100 sm:text-xs">
          <Layers className="size-4 shrink-0" />
          <span>Section {index + 1}</span>
        </div>
        <HeadingWithSource className="[&_h3]:text-2xl [&_h3]:font-semibold [&_h3]:text-white" source={section.title.source}>
          {section.title.value}
        </HeadingWithSource>
      </div>
      <div className="p-5">
        {section.estimatedDuration.value ? (
          <div className="flex items-center gap-2 text-sm text-slate-600">
            <Clock className="size-4 text-teal-700" />
            <SourcedInline value={section.estimatedDuration} />
          </div>
        ) : null}
        <div className="mt-4">
          <ProseBlock text={section.description} />
        </div>
      </div>
    </section>
  );
}

function LabBlock({ lab }: { lab: ModuleLab }) {
  return (
    <section className="overflow-hidden rounded-lg border border-emerald-200 bg-white shadow-sm">
      <div className="bg-emerald-700 px-5 py-4 text-white">
        <div className="mb-1 flex items-center gap-2 text-sm font-semibold text-emerald-100 sm:text-xs">
          <FlaskConical className="size-4 shrink-0" />
          <span>Lab</span>
        </div>
        <HeadingWithSource className="[&_h3]:text-2xl [&_h3]:font-semibold [&_h3]:text-white" source={lab.title.source}>
          {lab.title.value}
        </HeadingWithSource>
      </div>
      <div className="p-5">
        {lab.timeLimit.value ? (
          <div className="flex items-center gap-2 text-sm text-slate-600">
            <Clock className="size-4 text-emerald-700" />
            <SourcedInline value={lab.timeLimit} />
          </div>
        ) : null}
        <div className="mt-4">
          <ProseBlock text={lab.description} />
        </div>
        {lab.challenges.length ? (
          <div className="mt-5 space-y-3">
            {lab.challenges.map((challenge, index) => (
              <div className="rounded-md border border-slate-200 bg-slate-50 p-4" key={moduleChallengeKey(challenge, index)}>
                <HeadingWithSource className="[&_h4]:font-semibold [&_h4]:text-slate-950" level="h4" source={challenge.title.source}>
                  {challenge.title.value}
                </HeadingWithSource>
                <div className="mt-3 space-y-2">
                  {challenge.goals.map((goal, goalIndex) => (
                    <div className="flex items-start gap-2 text-sm" key={moduleGoalKey(goal, goalIndex)}>
                      <CheckCircle2 className="mt-0.5 size-4 text-emerald-700" />
                      <div>
                        <div className="flex items-center gap-1.5 font-medium text-slate-900">
                          <span>{goal.title.value}</span>
                          <SourceMarker source={goal.title.source} />
                        </div>
                        {goal.description.value ? <p className="text-slate-600">{goal.description.value}</p> : null}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : null}
      </div>
    </section>
  );
}

function QuizBlock({ quiz }: { quiz: ModuleQuiz }) {
  return (
    <section className="overflow-hidden rounded-lg border border-amber-200 bg-white shadow-sm">
      <div className="bg-amber-700 px-5 py-4 text-white">
        <div className="mb-1 flex items-center gap-2 text-sm font-semibold text-amber-100 sm:text-xs">
          <HelpCircle className="size-4 shrink-0" />
          <span>Quiz</span>
        </div>
        <HeadingWithSource className="[&_h3]:text-2xl [&_h3]:font-semibold [&_h3]:text-white" source={quiz.title.source}>
          {quiz.title.value}
        </HeadingWithSource>
      </div>
      <div className="p-5">
        <div className="flex flex-wrap gap-2">
          <SourcedInline
            className="rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-800"
            value={{ value: `${quiz.passingScore.value}% pass`, source: quiz.passingScore.source }}
          />
        </div>
        <div className="mt-4">
          <ProseBlock text={quiz.description} label="Prompt" />
        </div>
        <div className="mt-5 space-y-4">
          {quiz.questions.map((question, questionIndex) => (
            <div className="overflow-hidden rounded-md border border-amber-200 bg-white shadow-sm" key={moduleQuestionKey(question, questionIndex)}>
              <div className="bg-amber-600 px-4 py-3 text-white">
                <div className="mb-1 flex flex-wrap items-center gap-2 text-sm font-semibold text-amber-100 sm:text-xs">
                  <BadgeCheck className="size-4 shrink-0" />
                  <span>Question {questionIndex + 1}</span>
                  <SourcedInline
                    className="rounded-full bg-white/15 px-2 py-0.5 font-medium text-white ring-1 ring-white/20"
                    value={question.type}
                  />
                </div>
                <HeadingWithSource className="[&_h4]:text-xl [&_h4]:font-semibold [&_h4]:text-white" level="h4" source={question.question.source}>
                  {question.question.value}
                </HeadingWithSource>
              </div>
              <div className="grid gap-2 bg-amber-50/40 p-4 sm:grid-cols-2">
                {question.options.map((option, optionIndex) => (
                  <div
                    className="flex items-center gap-2 rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700"
                    key={`${moduleQuestionKey(question, questionIndex)}-option-${option.index || optionIndex}`}
                  >
                    {option.correct ? (
                      <CheckCircle2 className="size-4 text-emerald-700" />
                    ) : (
                      <Circle className="size-4 text-slate-300" />
                    )}
                    <span className="min-w-0 flex-1">{option.text.value}</span>
                    <SourceMarker source={option.text.source} />
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function EmptyState() {
  return (
    <div className="rounded-lg border border-dashed border-slate-300 bg-white p-10 text-center text-slate-600">
      No chapters found in this module.
    </div>
  );
}
