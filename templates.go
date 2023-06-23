package mutexvisualiser

const tmplGraph = `
digraph g{
rankdir="TB";
nodesep=0.5;
ranksep=0.25;
splines=line;
forcelabels=false;

// general
node [style=filled, color="black", 
    fontcolor="black", font="Consolas", fontsize="8pt" ];
edge [arrowhead=vee, color="black", penwidth=2,dir=none, fontsize="6pt"];

// graph
node [width=0.2, height=0.2,  label="", margin="0.11,0.055", shape=circle, penwidth=2, fillcolor="#FF8000"]

// timeline
t1[group="time", label="time",shape="square"];

// mutex
m1[group="mutex",label="mutex", shape="square"];

// main
g1_1[group="g1",label="main", shape="square"];

subgraph{
    rank="same"
    g1_1;
    m1;
}

{{.}}
}
`
const tmplCreateGBranch = `
g{{.GID}}_1[group="g{{ .GID }}", label="{{ .GID }}", shape="square"];
g{{.PGID}}_1->g{{.GID}}_1;





`

const tmplActionToMutex = `
t{{.NextTNo}}[shape="plaintext" group="time" label="{{.ActionTime}}"];
t{{.TNo}}->t{{.NextTNo}};
g{{.GID}}_{{.NextGNo}}[group="g{{.GID}}"];
g{{.GID}}_{{.GNo}}->g{{.GID}}_{{.NextGNo}};
m{{.NextMNo}} [group="mutex"];
m{{.MNo}}->m{{.NextMNo}};
subgraph{
    rank="same"
    m{{.NextMNo}};
    t{{.NextTNo}};
    g{{.GID}}_{{.NextGNo}};
}
g{{.GID}}_{{.NextGNo}}->m{{.NextMNo}}[label="{{.ActionType}}" dir="{{.ActionDir}}"];
`
const tmplLockMutex = `
t{{.NextTNo}}[shape="plaintext" group="time" label="{{.ActionTime}}"];
t{{.TNo}}->t{{.NextTNo}};
g{{.GID}}_{{.NextGNo}}[group="g{{.GID}}"];
g{{.GID}}_{{.GNo}}->g{{.GID}}_{{.NextGNo}}[label="waiting for lock...{{.Duration}}" color="{{.EdgeColor}}" dir="forward"];
m{{.NextMNo}} [group="mutex"];
m{{.MNo}}->m{{.NextMNo}};
subgraph{
    rank="same"
    m{{.NextMNo}};
    t{{.NextTNo}};
    g{{.GID}}_{{.NextGNo}};
}
g{{.GID}}_{{.NextGNo}}->m{{.NextMNo}}[label="lock" dir="forward"];
`
const tmplActionOnG = `
t{{.NextTNo}}[shape="plaintext" group="time" label="{{.ActionTime}}"];
t{{.TNo}}->t{{.NextTNo}};
g{{.GID}}_{{.NextGNo}}[group="g{{.GID}}"];
g{{.GID}}_{{.GNo}}->g{{.GID}}_{{.NextGNo}}[label="{{.ActionType}}" color="{{.EdgeColor}}"];	
subgraph{
    rank="same"
    t{{.NextTNo}};
    g{{.GID}}_{{.NextGNo}};
}
`
