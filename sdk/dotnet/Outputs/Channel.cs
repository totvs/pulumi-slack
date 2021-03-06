// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.Slack.Outputs
{

    [OutputType]
    public sealed class Channel
    {
        public readonly string Id;
        public readonly bool IsArchived;
        public readonly bool IsPrivate;
        public readonly ImmutableArray<string> Members;
        public readonly string Name;
        public readonly string? Purpose;
        public readonly string? Topic;

        [OutputConstructor]
        private Channel(
            string id,

            bool isArchived,

            bool isPrivate,

            ImmutableArray<string> members,

            string name,

            string? purpose,

            string? topic)
        {
            Id = id;
            IsArchived = isArchived;
            IsPrivate = isPrivate;
            Members = members;
            Name = name;
            Purpose = purpose;
            Topic = topic;
        }
    }
}
