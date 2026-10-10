function result = renderView(model, viewName, varargin)
%RENDERVIEW Render a named view as machine-readable diagram data.

    ports = 'minimal';
    artifact = [];
    artifactRequested = false;
    for i = 1:2:numel(varargin)
        if i == numel(varargin)
            opensysml.internal.raise('opensysml:argument', ...
                'renderView options must be name-value pairs');
        end
        if strcmp(varargin{i}, 'ports'), ports = char(varargin{i+1});
        elseif strcmp(varargin{i}, 'artifact'), artifact = varargin{i+1}; artifactRequested = true;
        else
            opensysml.internal.raise('opensysml:argument', ...
                sprintf('renderView has no option ''%s''', varargin{i}));
        end
    end
    if ~any(strcmp(ports, {'minimal', 'full'}))
        opensysml.internal.raise('opensysml:argument', ...
            'ports must be ''minimal'' or ''full''');
    end
    capability = {'render_view'};
    if artifactRequested
        if ~isstruct(artifact) || ~isscalar(artifact) || ~isfield(artifact, 'form')
            opensysml.internal.raise('opensysml:argument', 'artifact must be a scalar struct with a form');
        end
        allowed = {'form', 'palette', 'drawingStyle', 'overlay', 'direction', 'unplaced', 'width', 'linkTemplate'};
        if ~all(ismember(fieldnames(artifact), allowed))
            opensysml.internal.raise('opensysml:argument', 'unknown artifact option');
        end
        capability{end+1} = 'render_view_artifact';
    end
    model.connection.requireAll(capability);
    request = struct('modelHash', model.hash, 'view', char(viewName));
    if strcmp(ports, 'full'), request.ports = 'full'; end
    if artifactRequested, request.artifact = artifact; end
    answer = opensysml.call(model.connection, 'RenderView', request, capability);
    if artifactRequested && (~isfield(answer, 'artifact') || ~isstruct(answer.artifact))
        opensysml.internal.raise('opensysml:protocol', 'RenderView response omitted the requested artifact');
    end
    result = opensysml.RenderedView(answer);
end
