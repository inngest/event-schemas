export const Status = {
  OPEN: "open",
  CLOSED: "closed",
} as const;
export type Status = typeof Status[keyof typeof Status];

export interface Some {
  with: string;
};

export const Action = {
  PUSH: "push",
  PULL: "pull",
  REBASE: "rebase",
} as const;
export type Action = typeof Action[keyof typeof Action];

export type Mixed = string | number;

export interface Event {
  name: string;
  data: {
    action: Action;
    status: Status;
    number: number;
    static: "lol this is content";
    optionalStatic?: "some opt content";
    staticNumber: 1;
    staticBool?: true;
    enabled: boolean;
    numeric: number;
    mixed: Mixed;
  };
  allow: {
    with: string;
    included: boolean;
  };
};
