export interface SourceRef {
  file: string;
  property?: string;
}

export interface Sourced<T> {
  value: T;
  source?: SourceRef;
}

export interface ModulePreview {
  kind: "module";
  code: Sourced<string>;
  title: Sourced<string>;
  description: Sourced<string>;
  shortDescription: Sourced<string>;
  level: Sourced<string>;
  bannerImage: Sourced<string>;
  bannerVideo: Sourced<string>;
  stats: {
    chapters: number;
    sections: number;
    quizzes: number;
    labs: number;
  };
  chapters: ModuleChapter[];
}

export interface ModuleChapter {
  code: Sourced<string>;
  index: number;
  title: Sourced<string>;
  description: Sourced<string>;
  shortDescription: Sourced<string>;
  bannerImage: Sourced<string>;
  bannerVideo: Sourced<string>;
  isDraft: boolean;
  sections: ModuleSection[];
  labs: ModuleLab[];
  multipleChoiceAssessments: ModuleQuiz[];
}

export interface ModuleSection {
  code: Sourced<string>;
  index: number;
  title: Sourced<string>;
  description: Sourced<string>;
  shortDescription: Sourced<string>;
  thumbnailDescription: Sourced<string>;
  thumbnail: Sourced<string>;
  video: Sourced<string>;
  estimatedDuration: Sourced<string>;
}

export interface ModuleLab {
  code: Sourced<string>;
  index: number;
  title: Sourced<string>;
  description: Sourced<string>;
  timeLimit: Sourced<string>;
  starterImageUri: Sourced<string>;
  validatorImageUri: Sourced<string>;
  imageVersion: Sourced<string>;
  video: Sourced<string>;
  challenges: ModuleChallenge[];
}

export interface ModuleChallenge {
  code: Sourced<string>;
  index: number;
  title: Sourced<string>;
  description: Sourced<string>;
  successMessage: Sourced<string>;
  estimatedDuration: Sourced<string>;
  video: Sourced<string>;
  goals: ModuleGoal[];
}

export interface ModuleGoal {
  code: Sourced<string>;
  index: number;
  title: Sourced<string>;
  description: Sourced<string>;
}

export interface ModuleQuiz {
  code: Sourced<string>;
  index: number;
  title: Sourced<string>;
  description: Sourced<string>;
  passingScore: Sourced<number>;
  questions: ModuleQuestion[];
}

export interface ModuleQuestion {
  code: Sourced<string>;
  index: number;
  question: Sourced<string>;
  type: Sourced<string>;
  options: ModuleOption[];
}

export interface ModuleOption {
  index: number;
  text: Sourced<string>;
  correct: boolean;
}

export interface LabPreview {
  kind: "lab";
  format: string;
  code: Sourced<string>;
  title: Sourced<string>;
  description: Sourced<string>;
  timeLimit: Sourced<string>;
  progress: LabProgress;
  statusError?: string;
  runtime?: LabRuntime;
  challenges: LabChallenge[];
}

export interface LabRuntime {
  runId: string;
  systemNamespace: string;
  workspaceNamespace: string;
  registryUrl: string;
  registryUsername: string;
  registryToken: string;
}

export interface LabChallenge {
  code: Sourced<string>;
  title: Sourced<string>;
  description: Sourced<string>;
  successMessage: Sourced<string>;
  progress: LabProgress;
  goals: LabGoal[];
}

export interface LabGoal {
  code: Sourced<string>;
  title: Sourced<string>;
  description: Sourced<string>;
  progress: LabProgress;
}

export interface LabProgress {
  conditionType: string;
  state: "complete" | "incomplete" | "unknown";
  label: string;
  status?: string;
  reason?: string;
  message?: string;
}

export type PreviewPayload = ModulePreview | LabPreview;
