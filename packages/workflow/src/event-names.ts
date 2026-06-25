export class EventNames {
	public static guard(workflow: string): string {
		return `workflow.${workflow}.guard`;
	}

	public static guardNamed(workflow: string, transition: string): string {
		return `workflow.${workflow}.guard.${transition}`;
	}

	public static leave(workflow: string): string {
		return `workflow.${workflow}.leave`;
	}

	public static leavePlace(workflow: string, place: string): string {
		return `workflow.${workflow}.leave.${place}`;
	}

	public static transition(workflow: string): string {
		return `workflow.${workflow}.transition`;
	}

	public static transitionNamed(workflow: string, transition: string): string {
		return `workflow.${workflow}.transition.${transition}`;
	}

	public static enter(workflow: string): string {
		return `workflow.${workflow}.enter`;
	}

	public static enterPlace(workflow: string, place: string): string {
		return `workflow.${workflow}.enter.${place}`;
	}

	public static entered(workflow: string): string {
		return `workflow.${workflow}.entered`;
	}

	public static enteredPlace(workflow: string, place: string): string {
		return `workflow.${workflow}.entered.${place}`;
	}

	public static completed(workflow: string): string {
		return `workflow.${workflow}.completed`;
	}

	public static completedNamed(workflow: string, transition: string): string {
		return `workflow.${workflow}.completed.${transition}`;
	}

	public static announce(workflow: string): string {
		return `workflow.${workflow}.announce`;
	}

	public static announceNamed(workflow: string, transition: string): string {
		return `workflow.${workflow}.announce.${transition}`;
	}
}
